package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSecurityHeaders(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	want := map[string]string{
		"X-Content-Type-Options":     "nosniff",
		"X-Frame-Options":            "DENY",
		"Referrer-Policy":            "no-referrer",
		"Cross-Origin-Opener-Policy": "same-origin",
	}
	for k, v := range want {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Error("Content-Security-Policy header missing")
	}
}

func TestSecurityHeaders_AppliedByRouter(t *testing.T) {
	r := NewRouter(Deps{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("router did not apply security headers; got %v", rec.Header())
	}
}

func TestRateLimiter_EvictsIdle(t *testing.T) {
	l := newIPRateLimiter(60)
	l.limiterFor("1.2.3.4")
	if n := l.sweep(time.Now()); n != 1 {
		t.Errorf("recently-seen limiter should survive sweep, kept %d", n)
	}
	if n := l.sweep(time.Now().Add(2 * limiterTTL)); n != 0 {
		t.Errorf("idle limiter should be evicted, kept %d", n)
	}
}
