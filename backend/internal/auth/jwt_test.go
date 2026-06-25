package auth

import (
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const jwtSecret = "test-secret-that-is-at-least-32b-long!!"

func TestIssueAndParseAccess_RoundTrip(t *testing.T) {
	tm := NewTokenManager(jwtSecret, 15*time.Minute)

	token, err := tm.IssueAccess("user-123", "admin")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	claims, err := tm.ParseAccess(token)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want user-123", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want admin", claims.Role)
	}
}

func TestParseAccess_RejectsWrongSecret(t *testing.T) {
	issuer := NewTokenManager(jwtSecret, 15*time.Minute)
	verifier := NewTokenManager("a-completely-different-secret-key-32b!", 15*time.Minute)

	token, err := issuer.IssueAccess("u", "student")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := verifier.ParseAccess(token); err == nil {
		t.Error("expected error for token signed with a different secret")
	}
}

func TestParseAccess_RejectsExpired(t *testing.T) {
	tm := NewTokenManager(jwtSecret, -time.Minute) // already expired
	token, err := tm.IssueAccess("u", "student")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := tm.ParseAccess(token); err == nil {
		t.Error("expected error for an expired token")
	}
}

func TestParseAccess_RejectsGarbageAndNoneAlg(t *testing.T) {
	tm := NewTokenManager(jwtSecret, 15*time.Minute)

	if _, err := tm.ParseAccess("not.a.jwt"); err == nil {
		t.Error("expected error for malformed token")
	}

	// alg=none must be rejected (algorithm-confusion guard).
	none := jwt.NewWithClaims(jwt.SigningMethodNone, jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{Issuer: tokenIssuer, Subject: "u"},
	})
	noneStr, err := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}
	if _, err := tm.ParseAccess(noneStr); err == nil {
		t.Error("expected error for alg=none token")
	}
}
