package auth

import "testing"

func TestNewToken_UniqueAndHashed(t *testing.T) {
	tok1, hash1, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	tok2, hash2, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	if tok1 == "" || hash1 == "" {
		t.Fatal("token and hash must be non-empty")
	}
	if tok1 == tok2 {
		t.Error("two tokens should differ")
	}
	if hash1 == hash2 {
		t.Error("two token hashes should differ")
	}
	if tok1 == hash1 {
		t.Error("token must not equal its hash")
	}
	if HashToken(tok1) != hash1 {
		t.Error("HashToken is not deterministic for the same token")
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	if HashToken("abc") != HashToken("abc") {
		t.Error("HashToken should be deterministic")
	}
	if HashToken("abc") == HashToken("abd") {
		t.Error("different inputs should hash differently")
	}
	// SHA-256 hex is 64 characters.
	if got := len(HashToken("abc")); got != 64 {
		t.Errorf("hash length = %d, want 64", got)
	}
}
