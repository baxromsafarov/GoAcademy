package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// tokenBytes is the entropy of opaque tokens (email verification, password reset).
const tokenBytes = 32

// NewToken returns a cryptographically random URL-safe token and its SHA-256 hash
// (hex). The plaintext is delivered to the user (e.g. by email); only the hash is
// persisted, so a database leak does not expose usable tokens.
func NewToken() (token, hash string, err error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token), nil
}

// HashToken returns the hex-encoded SHA-256 of token, used to look up the stored
// hash. SHA-256 (not argon2) is appropriate here because tokens carry full entropy.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
