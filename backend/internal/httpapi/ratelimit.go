package httpapi

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
)

// limiterTTL is how long an idle per-IP limiter is kept before eviction.
const limiterTTL = 10 * time.Minute

type limiterEntry struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// ipRateLimiter keeps a token-bucket limiter per client IP. Idle limiters are
// evicted after limiterTTL by a background sweeper so the map cannot grow
// unbounded with the IP cardinality.
type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	limit    rate.Limit
	burst    int
}

func newIPRateLimiter(perMinute int) *ipRateLimiter {
	l := &ipRateLimiter{
		limiters: make(map[string]*limiterEntry),
		limit:    rate.Limit(float64(perMinute) / 60.0), // tokens per second
		burst:    perMinute,
	}
	go l.cleanupLoop()
	return l
}

func (l *ipRateLimiter) limiterFor(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.limiters[ip]
	if !ok {
		e = &limiterEntry{lim: rate.NewLimiter(l.limit, l.burst)}
		l.limiters[ip] = e
	}
	e.lastSeen = time.Now()
	return e.lim
}

// sweep removes limiters not seen since now-limiterTTL. Returns the number kept.
func (l *ipRateLimiter) sweep(now time.Time) int {
	cutoff := now.Add(-limiterTTL)
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, e := range l.limiters {
		if e.lastSeen.Before(cutoff) {
			delete(l.limiters, ip)
		}
	}
	return len(l.limiters)
}

func (l *ipRateLimiter) cleanupLoop() {
	t := time.NewTicker(limiterTTL)
	defer t.Stop()
	for range t.C {
		l.sweep(time.Now())
	}
}

// RateLimit returns middleware that allows at most perMinute requests per client
// IP (with a burst of perMinute), responding 429 when exceeded.
func RateLimit(perMinute int) func(http.Handler) http.Handler {
	limiter := newIPRateLimiter(perMinute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.limiterFor(clientIP(r)).Allow() {
				respond.Error(w, r, nil, apierr.RateLimited("too many requests, please slow down"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the host portion of the real TCP peer (RemoteAddr). It does
// not trust X-Forwarded-For / X-Real-IP (see the RealIP note in NewRouter), so a
// client cannot spoof its IP to dodge the per-IP rate limit.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
