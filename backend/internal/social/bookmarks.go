package social

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// Bookmark is a user's saved reference to a piece of content.
type Bookmark struct {
	ID          string
	ContentType string
	ContentID   string
	Title       string
	CreatedAt   time.Time
}

// BookmarksService manages a user's bookmarks.
type BookmarksService struct {
	queries *store.Queries
}

// NewBookmarksService wires the bookmarks service to the database.
func NewBookmarksService(pool *pgxpool.Pool) *BookmarksService {
	return &BookmarksService{queries: store.New(pool)}
}

// Add bookmarks content for the user. It is idempotent: re-adding the same
// content returns the existing bookmark rather than creating a duplicate.
func (s *BookmarksService) Add(ctx context.Context, userID, contentType, contentID string) (Bookmark, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return Bookmark{}, apierr.Unauthorized("invalid user")
	}
	details := map[string]string{}
	if !validContentType(contentType) {
		details["content_type"] = "must be one of: video, article, quiz, problem, project, track, cheatsheet, glossary"
	}
	cid, err := pgxutil.ParseUUID(contentID)
	if err != nil {
		details["content_id"] = "must be a valid uuid"
	}
	if len(details) > 0 {
		return Bookmark{}, apierr.Validation("invalid bookmark").WithDetails(details)
	}
	row, err := s.queries.CreateBookmark(ctx, store.CreateBookmarkParams{
		UserID: uid, ContentType: contentType, ContentID: cid,
	})
	if err != nil {
		return Bookmark{}, err
	}
	return toBookmark(row), nil
}

// Remove deletes a bookmark. Only the owner can remove it; anyone else (or a
// missing bookmark) gets not-found.
func (s *BookmarksService) Remove(ctx context.Context, userID, bookmarkID string) error {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return apierr.Unauthorized("invalid user")
	}
	bid, err := pgxutil.ParseUUID(bookmarkID)
	if err != nil {
		return apierr.NotFound("bookmark not found")
	}
	n, err := s.queries.DeleteBookmark(ctx, store.DeleteBookmarkParams{ID: bid, UserID: uid})
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("bookmark not found")
	}
	return nil
}

// List returns the user's bookmarks, newest first.
func (s *BookmarksService) List(ctx context.Context, userID string) ([]Bookmark, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return nil, apierr.Unauthorized("invalid user")
	}
	rows, err := s.queries.ListUserBookmarks(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]Bookmark, len(rows))
	for i, r := range rows {
		out[i] = Bookmark{
			ID:          pgxutil.UUIDString(r.ID),
			ContentType: r.ContentType,
			ContentID:   pgxutil.UUIDString(r.ContentID),
			Title:       r.Title,
			CreatedAt:   r.CreatedAt.Time,
		}
	}
	return out, nil
}

func toBookmark(b store.Bookmark) Bookmark {
	return Bookmark{
		ID:          pgxutil.UUIDString(b.ID),
		ContentType: b.ContentType,
		ContentID:   pgxutil.UUIDString(b.ContentID),
		CreatedAt:   b.CreatedAt.Time,
	}
}
