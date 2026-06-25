package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/store"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testTokens() *TokenManager {
	return NewTokenManager(jwtSecret, 15*time.Minute)
}

func testTTL() TTLConfig {
	return TTLConfig{EmailVerification: time.Hour, Refresh: time.Hour, PasswordReset: time.Hour}
}

// mailCapture records the last emails so tests can read the tokens.
type mailCapture struct {
	to         string
	token      string
	resetTo    string
	resetToken string
}

func (m *mailCapture) SendEmailVerification(_ context.Context, to, token string) error {
	m.to, m.token = to, token
	return nil
}

func (m *mailCapture) SendPasswordReset(_ context.Context, to, token string) error {
	m.resetTo, m.resetToken = to, token
	return nil
}

func TestRegisterInput_NormalizeAndValidate_OK(t *testing.T) {
	in := RegisterInput{Email: "  User@Example.com ", Password: "password123", DisplayName: "  Alice  ", Locale: ""}
	in.normalize()

	if in.Email != "user@example.com" {
		t.Errorf("email = %q, want normalized lowercase/trimmed", in.Email)
	}
	if in.DisplayName != "Alice" {
		t.Errorf("display_name = %q, want trimmed", in.DisplayName)
	}
	if in.Locale != store.LocaleEn {
		t.Errorf("locale = %q, want default en", in.Locale)
	}
	if err := in.validate(); err != nil {
		t.Errorf("validate returned error for valid input: %v", err)
	}
}

func TestRegisterInput_Validate_CollectsFieldErrors(t *testing.T) {
	in := RegisterInput{Email: "not-an-email", Password: "short", DisplayName: "", Locale: "xx"}
	in.normalize()

	err := in.validate()
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusBadRequest {
		t.Fatalf("expected 400 validation error, got %v", err)
	}
	details, ok := apiErr.Details.(map[string]string)
	if !ok {
		t.Fatalf("details not a map: %T", apiErr.Details)
	}
	for _, field := range []string{"email", "password", "display_name", "locale"} {
		if _, present := details[field]; !present {
			t.Errorf("expected a validation error for %q; got %v", field, details)
		}
	}
}

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run auth service integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestService_RegisterAndVerify_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	email := fmt.Sprintf("reg-%d@example.com", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })

	mc := &mailCapture{}
	svc := NewService(pool, mc, testTokens(), testLogger(), testTTL())

	user, err := svc.Register(ctx, RegisterInput{Email: email, Password: "password123", DisplayName: "Reg User"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Email != email {
		t.Errorf("email = %q, want %q", user.Email, email)
	}
	if user.EmailVerified {
		t.Error("new user should be unverified")
	}
	if user.Role != store.UserRoleStudent {
		t.Errorf("role = %q, want student", user.Role)
	}
	if mc.to != email || mc.token == "" {
		t.Fatalf("mailer not invoked correctly: to=%q token-empty=%v", mc.to, mc.token == "")
	}

	// Verify with the emailed token.
	if err := svc.VerifyEmail(ctx, mc.token); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}
	q := store.New(pool)
	u, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if !u.EmailVerified {
		t.Error("email should be verified after VerifyEmail")
	}

	// A reused token must be rejected.
	if err := svc.VerifyEmail(ctx, mc.token); err == nil {
		t.Error("reused token should be rejected")
	}
	// An unknown token must be rejected.
	if err := svc.VerifyEmail(ctx, "definitely-not-a-real-token"); err == nil {
		t.Error("invalid token should be rejected")
	}
}

func TestService_Register_DuplicateEmail_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	email := fmt.Sprintf("dup-%d@example.com", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })

	svc := NewService(pool, &mailCapture{}, testTokens(), testLogger(), testTTL())
	if _, err := svc.Register(ctx, RegisterInput{Email: email, Password: "password123", DisplayName: "A"}); err != nil {
		t.Fatalf("first register: %v", err)
	}

	_, err := svc.Register(ctx, RegisterInput{Email: email, Password: "password123", DisplayName: "B"})
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict {
		t.Fatalf("expected 409 conflict for duplicate email, got %v", err)
	}
}
