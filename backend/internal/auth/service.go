package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/mailer"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// Service implements authentication use cases. STAGE 2.2 covers registration and
// email verification; login/refresh/reset are added in later stages.
// TTLConfig holds the token lifetimes used by the auth service.
type TTLConfig struct {
	EmailVerification time.Duration
	Refresh           time.Duration
	PasswordReset     time.Duration
}

type Service struct {
	pool    *pgxpool.Pool
	queries *store.Queries
	mailer  mailer.Mailer
	tokens  *TokenManager
	logger  *slog.Logger
	ttl     TTLConfig
}

// NewService wires the auth service to its dependencies.
func NewService(pool *pgxpool.Pool, m mailer.Mailer, tokens *TokenManager, logger *slog.Logger, ttl TTLConfig) *Service {
	return &Service{
		pool:    pool,
		queries: store.New(pool),
		mailer:  m,
		tokens:  tokens,
		logger:  logger,
		ttl:     ttl,
	}
}

// RegisterInput is the validated input to Register.
type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
	Locale      store.Locale
}

func (in *RegisterInput) normalize() {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.Locale == "" {
		in.Locale = store.LocaleEn
	}
}

func (in RegisterInput) validate() error {
	details := map[string]string{}
	if addr, err := mail.ParseAddress(in.Email); err != nil || addr.Address != in.Email {
		details["email"] = "must be a valid email address"
	}
	if n := len(in.Password); n < 8 || n > 128 {
		details["password"] = "must be between 8 and 128 characters"
	}
	if rs := []rune(in.DisplayName); len(rs) == 0 || len(rs) > 100 {
		details["display_name"] = "must be 1..100 characters"
	}
	if !validLocale(in.Locale) {
		details["locale"] = "must be one of ru, en, uz, ja"
	}
	if len(details) > 0 {
		return apierr.Validation("validation failed").WithDetails(details)
	}
	return nil
}

func validLocale(l store.Locale) bool {
	switch l {
	case store.LocaleRu, store.LocaleEn, store.LocaleUz, store.LocaleJa:
		return true
	default:
		return false
	}
}

// Register creates a new student account (unverified), issues an email
// verification token (only its hash is stored) and asks the mailer to deliver it.
// User and token are created in one transaction.
func (s *Service) Register(ctx context.Context, in RegisterInput) (store.User, error) {
	in.normalize()
	if err := in.validate(); err != nil {
		return store.User{}, err
	}

	passwordHash, err := Hash(in.Password)
	if err != nil {
		return store.User{}, err
	}

	token, tokenHash, err := NewToken()
	if err != nil {
		return store.User{}, err
	}
	expiresAt := time.Now().Add(s.ttl.EmailVerification)

	var user store.User
	err = s.withTx(ctx, func(q *store.Queries) error {
		u, err := q.CreateUser(ctx, store.CreateUserParams{
			Email:        in.Email,
			PasswordHash: passwordHash,
			DisplayName:  in.DisplayName,
			Locale:       in.Locale,
		})
		if err != nil {
			if isUniqueViolation(err) {
				return apierr.Conflict("email already registered")
			}
			return err
		}
		if _, err := q.CreateEmailVerificationToken(ctx, store.CreateEmailVerificationTokenParams{
			UserID:    u.ID,
			TokenHash: tokenHash,
			ExpiresAt: pgxutil.Timestamptz(expiresAt),
		}); err != nil {
			return err
		}
		user = u
		return nil
	})
	if err != nil {
		return store.User{}, err
	}

	// Sending is best-effort: a mailer failure must not roll back a committed
	// registration. The user can request a new verification email later.
	if err := s.mailer.SendEmailVerification(ctx, user.Email, token); err != nil {
		s.logger.WarnContext(ctx, "failed to send verification email", "error", err, "user_id", pgxutil.UUIDString(user.ID))
	}
	return user, nil
}

// VerifyEmail consumes a verification token: it must exist, be unused and unexpired.
// On success the token is marked used and the user's email is flagged verified
// (both in one transaction). Invalid, used or expired tokens yield a 400.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return apierr.Validation("token is required")
	}
	tokenHash := HashToken(token)

	return s.withTx(ctx, func(q *store.Queries) error {
		t, err := q.GetEmailVerificationToken(ctx, tokenHash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apierr.Validation("invalid or expired verification token")
			}
			return err
		}
		if t.UsedAt.Valid {
			return apierr.Validation("verification token has already been used")
		}
		if t.ExpiresAt.Time.Before(time.Now()) {
			return apierr.Validation("invalid or expired verification token")
		}
		if err := q.MarkEmailVerificationTokenUsed(ctx, t.ID); err != nil {
			return err
		}
		return q.SetEmailVerified(ctx, t.UserID)
	})
}

// withTx runs fn inside a transaction, committing on success and rolling back on
// any error (including a panic via the deferred Rollback).
func (s *Service) withTx(ctx context.Context, fn func(*store.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
