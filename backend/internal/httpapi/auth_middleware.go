package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/goacademy/backend/internal/auth"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
)

type contextKey string

const (
	ctxUserID contextKey = "user_id"
	ctxRole   contextKey = "role"
)

// RequireAuth validates the Bearer access token and stores the user id and role
// in the request context. Missing, malformed, expired or invalid tokens get 401.
func RequireAuth(tokens *auth.TokenManager, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				respond.Error(w, r, logger, apierr.Unauthorized("missing or malformed Authorization header"))
				return
			}
			claims, err := tokens.ParseAccess(token)
			if err != nil {
				respond.Error(w, r, logger, apierr.Unauthorized("invalid or expired access token"))
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole allows the request only if the authenticated user has the given
// role. It must run after RequireAuth (which populates the role in context):
// a missing role yields 401, a mismatched role yields 403.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actual, ok := RoleFromContext(r.Context())
			if !ok {
				respond.Error(w, r, nil, apierr.Unauthorized("authentication required"))
				return
			}
			if actual != role {
				respond.Error(w, r, nil, apierr.Forbidden("insufficient permissions"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth populates the user id and role in context when a valid Bearer
// token is present, but never rejects the request. It is for endpoints that are
// public yet behave differently for an authenticated (e.g. admin) caller — such
// as content lists that reveal hidden items only to admins.
func OptionalAuth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token, ok := bearerToken(r.Header.Get("Authorization")); ok {
				if claims, err := tokens.ParseAccess(token); err == nil {
					ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
					ctx = context.WithValue(ctx, ctxRole, claims.Role)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

// UserIDFromContext returns the authenticated user id set by RequireAuth.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxUserID).(string)
	return v, ok && v != ""
}

// RoleFromContext returns the authenticated user's role set by RequireAuth.
func RoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxRole).(string)
	return v, ok && v != ""
}
