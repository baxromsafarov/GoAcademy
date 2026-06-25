// Package auth provides password hashing and (in later stages) tokens, sessions
// and authorization for GoAcademy.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Errors returned when decoding an encoded hash.
var (
	ErrInvalidHash         = errors.New("argon2: invalid encoded hash")
	ErrIncompatibleVersion = errors.New("argon2: incompatible argon2 version")
)

// argon2Params are the cost parameters baked into each encoded hash, so existing
// hashes keep verifying even if defaults change later.
type argon2Params struct {
	memory      uint32 // KiB
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// defaultParams follow OWASP argon2id guidance (64 MiB, t=3, p=2).
var defaultParams = argon2Params{
	memory:      64 * 1024,
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

// Hash returns an argon2id PHC-formatted hash of password
// ($argon2id$v=19$m=...,t=...,p=...$salt$key). A fresh random salt is used each
// call, so identical passwords produce different hashes.
func Hash(password string) (string, error) {
	return hashWithParams(password, defaultParams)
}

func hashWithParams(password string, p argon2Params) (string, error) {
	salt := make([]byte, p.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("argon2: generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, uint8(p.parallelism), p.keyLength)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.iterations, p.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return encoded, nil
}

// Verify reports whether password matches encodedHash. It returns false (not an
// error) for a non-matching password, and an error only for a malformed hash.
func Verify(password, encodedHash string) (bool, error) {
	p, salt, key, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	other := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, uint8(p.parallelism), p.keyLength)

	if subtle.ConstantTimeEq(int32(len(key)), int32(len(other))) == 0 {
		return false, nil
	}
	return subtle.ConstantTimeCompare(key, other) == 1, nil
}

func decodeHash(encoded string) (argon2Params, []byte, []byte, error) {
	// Expected: ["", "argon2id", "v=19", "m=..,t=..,p=..", saltB64, keyB64]
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argon2Params{}, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return argon2Params{}, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return argon2Params{}, nil, nil, ErrIncompatibleVersion
	}

	var p argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism); err != nil {
		return argon2Params{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argon2Params{}, nil, nil, ErrInvalidHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argon2Params{}, nil, nil, ErrInvalidHash
	}

	p.saltLength = uint32(len(salt))
	p.keyLength = uint32(len(key))
	return p, salt, key, nil
}
