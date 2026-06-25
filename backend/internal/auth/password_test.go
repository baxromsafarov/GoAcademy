package auth

import (
	"strings"
	"testing"
)

func TestHashAndVerify_RoundTrip(t *testing.T) {
	const password = "correct horse battery staple"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := Verify(password, hash)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Error("Verify returned false for the correct password")
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	hash, err := Hash("right-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	ok, err := Verify("wrong-password", hash)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if ok {
		t.Error("Verify returned true for an incorrect password")
	}
}

func TestHash_UsesRandomSalt(t *testing.T) {
	h1, err := Hash("same-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	h2, err := Hash("same-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if h1 == h2 {
		t.Error("two hashes of the same password are identical; salt is not random")
	}
	// Both must still verify.
	for i, h := range []string{h1, h2} {
		ok, err := Verify("same-password", h)
		if err != nil || !ok {
			t.Errorf("hash %d failed to verify (ok=%v err=%v)", i, ok, err)
		}
	}
}

func TestHash_EncodingFormat(t *testing.T) {
	hash, err := Hash("x")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("unexpected hash prefix: %q", hash)
	}
	if n := strings.Count(hash, "$"); n != 5 {
		t.Errorf("expected 5 '$' separators, got %d in %q", n, hash)
	}
}

func TestVerify_MalformedHash(t *testing.T) {
	cases := []string{
		"",
		"not-a-hash",
		"$argon2id$v=19$m=65536,t=3,p=2$only-salt",    // missing key segment
		"$bcrypt$v=19$m=1,t=1,p=1$c2FsdA$a2V5",        // wrong algorithm
		"$argon2id$v=999$m=65536,t=3,p=2$c2FsdA$a2V5", // incompatible version
	}
	for _, c := range cases {
		if _, err := Verify("x", c); err == nil {
			t.Errorf("Verify(%q) = nil error, want error", c)
		}
	}
}
