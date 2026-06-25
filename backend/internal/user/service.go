// Package user implements the student profile use cases (read and edit "me").
package user

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/platform/storage"
	"github.com/goacademy/backend/internal/store"
)

// Service reads and updates user profiles.
type Service struct {
	queries *store.Queries
	storage storage.Storage
}

// NewService wires the user service to the database and blob storage.
func NewService(pool *pgxpool.Pool, st storage.Storage) *Service {
	return &Service{queries: store.New(pool), storage: st}
}

// GetByID returns the profile for the given user id.
func (s *Service) GetByID(ctx context.Context, id string) (store.User, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.User{}, apierr.NotFound("user not found")
	}
	user, err := s.queries.GetUserByID(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.User{}, apierr.NotFound("user not found")
	}
	if err != nil {
		return store.User{}, err
	}
	return user, nil
}

// UpdateInput is a partial profile update: a nil field is left unchanged.
type UpdateInput struct {
	DisplayName *string
	Bio         *string
	Location    *string
	Locale      *string
	IsPublic    *bool
}

var validLocales = map[string]struct{}{
	string(store.LocaleRu): {},
	string(store.LocaleEn): {},
	string(store.LocaleUz): {},
	string(store.LocaleJa): {},
}

func (in *UpdateInput) validate() error {
	details := map[string]string{}
	if in.DisplayName != nil {
		*in.DisplayName = strings.TrimSpace(*in.DisplayName)
		if n := len([]rune(*in.DisplayName)); n < 1 || n > 100 {
			details["display_name"] = "must be 1..100 characters"
		}
	}
	if in.Bio != nil && len([]rune(*in.Bio)) > 500 {
		details["bio"] = "must be at most 500 characters"
	}
	if in.Location != nil && len([]rune(*in.Location)) > 100 {
		details["location"] = "must be at most 100 characters"
	}
	if in.Locale != nil {
		if _, ok := validLocales[*in.Locale]; !ok {
			details["locale"] = "must be one of ru, en, uz, ja"
		}
	}
	if len(details) > 0 {
		return apierr.Validation("validation failed").WithDetails(details)
	}
	return nil
}

// Update applies a partial profile update and returns the updated user.
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (store.User, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.User{}, apierr.NotFound("user not found")
	}
	if err := in.validate(); err != nil {
		return store.User{}, err
	}

	params := store.UpdateUserProfileParams{ID: uid}
	if in.DisplayName != nil {
		params.DisplayName = pgtype.Text{String: *in.DisplayName, Valid: true}
	}
	if in.Bio != nil {
		params.Bio = pgtype.Text{String: *in.Bio, Valid: true}
	}
	if in.Location != nil {
		params.Location = pgtype.Text{String: *in.Location, Valid: true}
	}
	if in.Locale != nil {
		params.Locale = store.NullLocale{Locale: store.Locale(*in.Locale), Valid: true}
	}
	if in.IsPublic != nil {
		params.IsPublic = pgtype.Bool{Bool: *in.IsPublic, Valid: true}
	}

	user, err := s.queries.UpdateUserProfile(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.User{}, apierr.NotFound("user not found")
	}
	if err != nil {
		return store.User{}, err
	}
	return user, nil
}

// SetAvatar stores the avatar image under a stable per-user key and updates the
// user's avatar_url. ext is the file extension (including the dot), e.g. ".png".
func (s *Service) SetAvatar(ctx context.Context, id, ext string, data io.Reader) (store.User, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.User{}, apierr.NotFound("user not found")
	}

	url, err := s.storage.Save(ctx, "avatars/"+id+ext, data)
	if err != nil {
		return store.User{}, err
	}

	user, err := s.queries.UpdateAvatarURL(ctx, store.UpdateAvatarURLParams{
		AvatarUrl: pgtype.Text{String: url, Valid: true},
		ID:        uid,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.User{}, apierr.NotFound("user not found")
	}
	if err != nil {
		return store.User{}, err
	}
	return user, nil
}
