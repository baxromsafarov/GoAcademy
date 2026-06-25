package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// RequestPasswordReset issues a reset token for the account with the given email
// and asks the mailer to deliver it. To avoid account enumeration it always
// succeeds for the caller, whether or not the email is registered.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // unknown email: stay silent
	}
	if err != nil {
		return err
	}

	token, tokenHash, err := NewToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(s.ttl.PasswordReset)
	if _, err := s.queries.CreatePasswordResetToken(ctx, store.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgxutil.Timestamptz(expiresAt),
	}); err != nil {
		return err
	}

	if err := s.mailer.SendPasswordReset(ctx, user.Email, token); err != nil {
		s.logger.WarnContext(ctx, "failed to send password reset email", "error", err, "user_id", pgxutil.UUIDString(user.ID))
	}
	return nil
}

// ResetPassword consumes a reset token and sets a new password. It revokes all of
// the user's refresh sessions so any existing logins are invalidated. The token
// must exist, be unused and unexpired; otherwise a 400 is returned.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if strings.TrimSpace(token) == "" {
		return apierr.Validation("token is required")
	}
	if n := len(newPassword); n < 8 || n > 128 {
		return apierr.Validation("password must be between 8 and 128 characters")
	}

	tokenHash := HashToken(token)
	newHash, err := Hash(newPassword)
	if err != nil {
		return err
	}

	return s.withTx(ctx, func(q *store.Queries) error {
		t, err := q.GetPasswordResetToken(ctx, tokenHash)
		if errors.Is(err, pgx.ErrNoRows) {
			return apierr.Validation("invalid or expired reset token")
		}
		if err != nil {
			return err
		}
		if t.UsedAt.Valid {
			return apierr.Validation("reset token has already been used")
		}
		if t.ExpiresAt.Time.Before(time.Now()) {
			return apierr.Validation("invalid or expired reset token")
		}

		if err := q.MarkPasswordResetTokenUsed(ctx, t.ID); err != nil {
			return err
		}
		if err := q.UpdatePasswordHash(ctx, store.UpdatePasswordHashParams{
			PasswordHash: newHash,
			ID:           t.UserID,
		}); err != nil {
			return err
		}
		// Invalidate every existing session for this user.
		return q.RevokeAllUserRefreshSessions(ctx, t.UserID)
	})
}
