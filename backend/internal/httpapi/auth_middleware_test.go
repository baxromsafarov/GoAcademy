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

func TestRequireAuth(t *testing.T) {
	tokens := auth.NewTokenManager("middleware-test-secret-at-least-32bytes!", 15*time.Minute)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	protected := RequireAuth(tokens, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected user id in context")
		}
		role, _ := RoleFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(uid + ":" + role))
	}))

	validToken, err := tokens.IssueAccess("user-9", "admin")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	cases := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"valid", "Bearer " + validToken, http.StatusOK},
		{"missing header", "", http.StatusUnauthorized},
		{"wrong scheme", "Token " + validToken, http.StatusUnauthorized},
		{"garbage token", "Bearer not.a.jwt", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusOK && rec.Body.String() != "user-9:admin" {
				t.Errorf("body = %q, want %q", rec.Body.String(), "user-9:admin")
			}
		})
	}
}
