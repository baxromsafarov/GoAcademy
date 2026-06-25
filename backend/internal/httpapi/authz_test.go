package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goacademy/backend/internal/auth"
)

func TestRequireRole(t *testing.T) {
	tokens := auth.NewTokenManager("role-test-secret-at-least-32-bytes-ok!", 15*time.Minute)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := RequireAuth(tokens, logger)(RequireRole("admin")(ok))

	adminToken, _ := tokens.IssueAccess("u1", "admin")
	studentToken, _ := tokens.IssueAccess("u2", "student")

	cases := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"admin allowed", "Bearer " + adminToken, http.StatusOK},
		{"student forbidden", "Bearer " + studentToken, http.StatusForbidden},
		{"no token unauthorized", "", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := RateLimit(2)(ok) // burst of 2 per IP

	call := func(ip string) int {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = ip + ":12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if call("1.2.3.4") != http.StatusOK || call("1.2.3.4") != http.StatusOK {
		t.Fatal("first two requests from an IP should pass")
	}
	if got := call("1.2.3.4"); got != http.StatusTooManyRequests {
		t.Errorf("third request status = %d, want 429", got)
	}
	if got := call("5.6.7.8"); got != http.StatusOK {
		t.Errorf("a different IP should not be limited, got %d", got)
	}
}
