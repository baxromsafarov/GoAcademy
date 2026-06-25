package auth

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when an access token is missing, malformed, expired
// or signed incorrectly.
var ErrInvalidToken = errors.New("invalid access token")

const tokenIssuer = "goacademy"

// AccessClaims is the validated payload of an access token.
type AccessClaims struct {
	UserID string
	Role   string
}

// TokenManager issues and verifies HS256 access tokens.
type TokenManager struct {
	secret    []byte
	accessTTL time.Duration
}

// NewTokenManager builds a TokenManager from the signing secret and access TTL.
func NewTokenManager(secret string, accessTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL}
}

// jwtClaims is the on-the-wire claim set.
type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// IssueAccess returns a signed access token for the given user and role.
func (tm *TokenManager) IssueAccess(userID, role string) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secret)
}

// ParseAccess verifies tokenString and returns its claims. Any problem
// (bad signature, wrong algorithm, expiry, malformed) yields ErrInvalidToken.
func (tm *TokenManager) ParseAccess(tokenString string) (AccessClaims, error) {
	var claims jwtClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return tm.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(tokenIssuer))
	if err != nil || !token.Valid {
		return AccessClaims{}, ErrInvalidToken
	}
	return AccessClaims{UserID: claims.Subject, Role: claims.Role}, nil
}
