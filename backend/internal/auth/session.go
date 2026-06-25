package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

const maxUserAgentLen = 512

// TokenPair is the result of a successful login or refresh.
type TokenPair struct {
	User             store.User
	AccessToken      string
	RefreshToken     string // plaintext refresh token, delivered as an httpOnly cookie
	RefreshExpiresAt time.Time
}

// Login verifies credentials and starts a new session (a fresh rotation family).
// Wrong email/password yield a generic 401; a blocked account yields 403.
func (s *Service) Login(ctx context.Context, email, password, userAgent string) (TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return TokenPair{}, apierr.Unauthorized("invalid email or password")
	}
	if err != nil {
		return TokenPair{}, err
	}
	if user.IsBlocked {
		return TokenPair{}, apierr.Forbidden("account is blocked")
	}

	ok, err := Verify(password, user.PasswordHash)
	if err != nil {
		return TokenPair{}, err
	}
	if !ok {
		return TokenPair{}, apierr.Unauthorized("invalid email or password")
	}

	familyID, err := pgxutil.NewUUID()
	if err != nil {
		return TokenPair{}, err
	}
	return s.issueSession(ctx, s.queries, user, familyID, userAgent)
}

// Refresh rotates a refresh token: it must be present, unrevoked and unexpired.
// A revoked token presented again is treated as a leak — the whole family is
// revoked (reuse detection). On success the old token is revoked and a new one
// (plus a new access token) is issued within the same family.
func (s *Service) Refresh(ctx context.Context, refreshToken, userAgent string) (TokenPair, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return TokenPair{}, apierr.Unauthorized("invalid refresh token")
	}
	tokenHash := HashToken(refreshToken)

	sess, err := s.queries.GetRefreshSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return TokenPair{}, apierr.Unauthorized("invalid refresh token")
	}
	if err != nil {
		return TokenPair{}, err
	}

	// Revocations below must persist even though we return an error to the caller,
	// so they run on the pool (each its own committed statement), not inside the
	// rotation transaction (which would roll them back).
	if sess.RevokedAt.Valid {
		_ = s.queries.RevokeRefreshFamily(ctx, sess.FamilyID) // reuse detected: kill the lineage
		return TokenPair{}, apierr.Unauthorized("invalid refresh token")
	}
	if sess.ExpiresAt.Time.Before(time.Now()) {
		_ = s.queries.RevokeRefreshSession(ctx, sess.ID)
		return TokenPair{}, apierr.Unauthorized("invalid refresh token")
	}

	user, err := s.queries.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return TokenPair{}, err
	}
	if user.IsBlocked {
		_ = s.queries.RevokeRefreshFamily(ctx, sess.FamilyID)
		return TokenPair{}, apierr.Forbidden("account is blocked")
	}

	// Rotation must be atomic: revoke the presented token and issue its successor
	// in the same family together.
	var pair TokenPair
	err = s.withTx(ctx, func(q *store.Queries) error {
		if err := q.RevokeRefreshSession(ctx, sess.ID); err != nil {
			return err
		}
		p, err := s.issueSession(ctx, q, user, sess.FamilyID, userAgent)
		if err != nil {
			return err
		}
		pair = p
		return nil
	})
	return pair, err
}

// Logout revokes the session for the presented refresh token. It is idempotent:
// an unknown or empty token is a no-op (no error).
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	sess, err := s.queries.GetRefreshSessionByTokenHash(ctx, HashToken(refreshToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.queries.RevokeRefreshSession(ctx, sess.ID)
}

// issueSession persists a new refresh session and mints the access token.
func (s *Service) issueSession(ctx context.Context, q *store.Queries, user store.User, familyID pgtype.UUID, userAgent string) (TokenPair, error) {
	if len(userAgent) > maxUserAgentLen {
		userAgent = userAgent[:maxUserAgentLen]
	}

	refreshToken, refreshHash, err := NewToken()
	if err != nil {
		return TokenPair{}, err
	}
	expiresAt := time.Now().Add(s.ttl.Refresh)

	if _, err := q.CreateRefreshSession(ctx, store.CreateRefreshSessionParams{
		UserID:    user.ID,
		FamilyID:  familyID,
		TokenHash: refreshHash,
		UserAgent: userAgent,
		ExpiresAt: pgxutil.Timestamptz(expiresAt),
	}); err != nil {
		return TokenPair{}, err
	}

	access, err := s.tokens.IssueAccess(pgxutil.UUIDString(user.ID), string(user.Role))
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		User:             user,
		AccessToken:      access,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: expiresAt,
	}, nil
}
