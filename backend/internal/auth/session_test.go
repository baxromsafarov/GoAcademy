package auth

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestService_LoginRefreshLogout_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	email := fmt.Sprintf("sess-%d@example.com", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })

	svc := NewService(pool, &mailCapture{}, testTokens(), testLogger(), testTTL())
	if _, err := svc.Register(ctx, RegisterInput{Email: email, Password: "password123", DisplayName: "Session"}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// --- login ---
	pair, err := svc.Login(ctx, email, "password123", "test-agent")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("login should return both access and refresh tokens")
	}
	claims, err := testTokens().ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatalf("access token should be valid: %v", err)
	}
	if claims.Role != "student" {
		t.Errorf("role claim = %q, want student", claims.Role)
	}

	// --- wrong password ---
	if _, err := svc.Login(ctx, email, "wrong-password", "ua"); err == nil {
		t.Error("login with wrong password should fail")
	}

	// --- refresh rotates the token ---
	pair2, err := svc.Refresh(ctx, pair.RefreshToken, "ua")
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if pair2.RefreshToken == pair.RefreshToken {
		t.Error("refresh must rotate the refresh token")
	}

	// --- reuse detection: presenting the old (rotated) token must fail ... ---
	if _, err := svc.Refresh(ctx, pair.RefreshToken, "ua"); err == nil {
		t.Error("reusing a rotated refresh token should fail")
	}
	// ... and must revoke the whole family, invalidating the latest token too.
	if _, err := svc.Refresh(ctx, pair2.RefreshToken, "ua"); err == nil {
		t.Error("after reuse detection the entire family should be revoked")
	}

	// --- logout revokes the session ---
	pair3, err := svc.Login(ctx, email, "password123", "ua")
	if err != nil {
		t.Fatalf("Login (3): %v", err)
	}
	if err := svc.Logout(ctx, pair3.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, err := svc.Refresh(ctx, pair3.RefreshToken, "ua"); err == nil {
		t.Error("refresh after logout should fail")
	}
}
