package auth

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestService_PasswordReset_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	email := fmt.Sprintf("reset-%d@example.com", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })

	mc := &mailCapture{}
	svc := NewService(pool, mc, testTokens(), testLogger(), testTTL())
	if _, err := svc.Register(ctx, RegisterInput{Email: email, Password: "oldpassword1", DisplayName: "Reset"}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Create a session that should be invalidated by the reset.
	pair, err := svc.Login(ctx, email, "oldpassword1", "ua")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// Unknown email: silent success, no token issued (no enumeration).
	if err := svc.RequestPasswordReset(ctx, "unknown-"+email); err != nil {
		t.Fatalf("RequestPasswordReset(unknown): %v", err)
	}
	if mc.resetToken != "" {
		t.Fatal("no reset token should be issued for an unknown email")
	}

	// Real email: a token is issued and "sent".
	if err := svc.RequestPasswordReset(ctx, email); err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	if mc.resetToken == "" || mc.resetTo != email {
		t.Fatalf("reset email not captured: to=%q token-empty=%v", mc.resetTo, mc.resetToken == "")
	}

	// Reset the password.
	if err := svc.ResetPassword(ctx, mc.resetToken, "newpassword2"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// Old password no longer works; the new one does.
	if _, err := svc.Login(ctx, email, "oldpassword1", "ua"); err == nil {
		t.Error("login with the old password should fail after reset")
	}
	if _, err := svc.Login(ctx, email, "newpassword2", "ua"); err != nil {
		t.Errorf("login with the new password should succeed: %v", err)
	}

	// The pre-reset session must be revoked.
	if _, err := svc.Refresh(ctx, pair.RefreshToken, "ua"); err == nil {
		t.Error("existing sessions should be revoked after a password reset")
	}

	// The reset token is single-use; reuse and garbage are rejected.
	if err := svc.ResetPassword(ctx, mc.resetToken, "another123"); err == nil {
		t.Error("a used reset token should be rejected")
	}
	if err := svc.ResetPassword(ctx, "definitely-not-a-real-token", "another123"); err == nil {
		t.Error("an invalid reset token should be rejected")
	}
}
