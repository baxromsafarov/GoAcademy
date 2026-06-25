package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

const (
	defaultUsersLimit = 20
	maxUsersLimit     = 100
)

// UserList is a paginated page of users.
type UserList struct {
	Items  []store.User
	Total  int64
	Limit  int
	Offset int
}

// ListUsers returns a page of users, optionally filtered by a search term over
// email and display name.
func (s *Service) ListUsers(ctx context.Context, q string, limit, offset int) (UserList, error) {
	if limit <= 0 {
		limit = defaultUsersLimit
	}
	if limit > maxUsersLimit {
		limit = maxUsersLimit
	}
	if offset < 0 {
		offset = 0
	}
	var search pgtype.Text
	if s := strings.TrimSpace(q); s != "" {
		search = pgtype.Text{String: s, Valid: true}
	}

	items, err := s.queries.ListUsers(ctx, store.ListUsersParams{Q: search, Lim: int32(limit), Off: int32(offset)})
	if err != nil {
		return UserList{}, err
	}
	total, err := s.queries.CountUsers(ctx, search)
	if err != nil {
		return UserList{}, err
	}
	return UserList{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

// UpdateUserInput carries the optional admin changes to a user.
type UpdateUserInput struct {
	Role      *string
	IsBlocked *bool
}

// UpdateUser changes a user's role and/or block status. Admins cannot demote or
// block themselves (so they can't accidentally lock themselves out).
func (s *Service) UpdateUser(ctx context.Context, actingUserID, targetID string, in UpdateUserInput) (store.User, error) {
	tid, err := pgxutil.ParseUUID(targetID)
	if err != nil {
		return store.User{}, apierr.NotFound("user not found")
	}

	if isSelf(actingUserID, tid) {
		if (in.Role != nil && *in.Role != "admin") || (in.IsBlocked != nil && *in.IsBlocked) {
			return store.User{}, apierr.Forbidden("you cannot demote or block yourself")
		}
	}
	if in.Role != nil && *in.Role != "student" && *in.Role != "admin" {
		return store.User{}, apierr.Validation("invalid role").WithDetails(map[string]string{"role": "must be student or admin"})
	}

	current, err := s.queries.GetUserByID(ctx, tid)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.User{}, apierr.NotFound("user not found")
	}
	if err != nil {
		return store.User{}, err
	}

	role := current.Role
	if in.Role != nil {
		role = store.UserRole(*in.Role)
	}
	blocked := current.IsBlocked
	if in.IsBlocked != nil {
		blocked = *in.IsBlocked
	}
	return s.queries.AdminUpdateUser(ctx, store.AdminUpdateUserParams{ID: tid, Role: role, IsBlocked: blocked})
}

// isSelf reports whether actingUserID refers to the same user as target.
func isSelf(actingUserID string, target pgtype.UUID) bool {
	aid, err := pgxutil.ParseUUID(actingUserID)
	if err != nil {
		return false
	}
	return pgxutil.UUIDString(aid) == pgxutil.UUIDString(target)
}
