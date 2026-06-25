// Package progress records a user's progress through content (videos, ...).
package progress

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// videoCompletionPercent is the watched-percent threshold for auto-completion.
const videoCompletionPercent = 90

// Service records content progress.
type Service struct {
	queries  *store.Queries
	activity activity.Recorder
}

// NewService wires the progress service to the database and the activity recorder.
func NewService(pool *pgxpool.Pool, rec activity.Recorder) *Service {
	return &Service{queries: store.New(pool), activity: rec}
}

// VideoProgressInput is a single progress report for a video.
type VideoProgressInput struct {
	Percent   int
	Position  int
	Completed *bool // explicit manual "mark as watched"
}

func (in VideoProgressInput) validate() error {
	details := map[string]string{}
	if in.Percent < 0 || in.Percent > 100 {
		details["percent"] = "must be between 0 and 100"
	}
	if in.Position < 0 {
		details["position"] = "must be >= 0"
	}
	if len(details) > 0 {
		return apierr.Validation("invalid progress").WithDetails(details)
	}
	return nil
}

// RecordVideoProgress upserts a user's progress for a video. watched_percent
// never decreases and completed is sticky; auto-completes at >=90% (or when the
// caller marks it manually). It records a "video_completed" activity event the
// first time the video becomes completed.
func (s *Service) RecordVideoProgress(ctx context.Context, userID, videoID string, in VideoProgressInput) (store.VideoProgress, error) {
	if err := in.validate(); err != nil {
		return store.VideoProgress{}, err
	}
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return store.VideoProgress{}, apierr.Unauthorized("invalid user")
	}
	vid, err := pgxutil.ParseUUID(videoID)
	if err != nil {
		return store.VideoProgress{}, apierr.NotFound("video not found")
	}

	// Determine whether it was already completed, so the activity fires only once.
	wasCompleted := false
	existing, err := s.queries.GetVideoProgress(ctx, store.GetVideoProgressParams{UserID: uid, VideoID: vid})
	switch {
	case err == nil:
		wasCompleted = existing.Completed
	case errors.Is(err, pgx.ErrNoRows):
		// no prior progress
	default:
		return store.VideoProgress{}, err
	}

	completed := in.Percent >= videoCompletionPercent
	if in.Completed != nil && *in.Completed {
		completed = true
	}

	row, err := s.queries.UpsertVideoProgress(ctx, store.UpsertVideoProgressParams{
		UserID:              uid,
		VideoID:             vid,
		WatchedPercent:      int32(in.Percent),
		LastPositionSeconds: int32(in.Position),
		Completed:           completed,
	})
	if isForeignKeyViolation(err) {
		return store.VideoProgress{}, apierr.NotFound("video not found")
	}
	if err != nil {
		return store.VideoProgress{}, err
	}

	if !wasCompleted && row.Completed {
		// Best-effort: a recorder failure must not fail the progress write.
		_ = s.activity.Record(ctx, activity.Event{
			UserID:  userID,
			Type:    "video_completed",
			RefType: "video",
			RefID:   videoID,
		})
	}
	return row, nil
}

// GetVideoProgress returns the user's saved progress for a video. When there is
// no saved progress it returns a zero-valued record (percent 0, not completed),
// so callers can read a position to resume from without special-casing absence.
func (s *Service) GetVideoProgress(ctx context.Context, userID, videoID string) (store.VideoProgress, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return store.VideoProgress{}, apierr.Unauthorized("invalid user")
	}
	vid, err := pgxutil.ParseUUID(videoID)
	if err != nil {
		return store.VideoProgress{}, apierr.NotFound("video not found")
	}
	row, err := s.queries.GetVideoProgress(ctx, store.GetVideoProgressParams{UserID: uid, VideoID: vid})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.VideoProgress{UserID: uid, VideoID: vid}, nil
	}
	if err != nil {
		return store.VideoProgress{}, err
	}
	return row, nil
}

// MarkArticleRead records that the user has read the article identified by slug.
// It is idempotent (a repeat is a no-op) and records an "article_read" activity
// event only on the first read.
func (s *Service) MarkArticleRead(ctx context.Context, userID, slug string) (store.ArticleRead, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return store.ArticleRead{}, apierr.Unauthorized("invalid user")
	}

	article, err := s.resolveArticle(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.ArticleRead{}, apierr.NotFound("article not found")
	}
	if err != nil {
		return store.ArticleRead{}, err
	}

	read, err := s.queries.MarkArticleRead(ctx, store.MarkArticleReadParams{UserID: uid, ArticleID: article.ID})
	switch {
	case err == nil:
		// First read: record the activity.
		_ = s.activity.Record(ctx, activity.Event{
			UserID:  userID,
			Type:    "article_read",
			RefType: "article",
			RefID:   pgxutil.UUIDString(article.ID),
		})
		return read, nil
	case errors.Is(err, pgx.ErrNoRows):
		// Already read: return the existing record without a new activity.
		read, err = s.queries.GetArticleRead(ctx, store.GetArticleReadParams{UserID: uid, ArticleID: article.ID})
		if err != nil {
			return store.ArticleRead{}, err
		}
		return read, nil
	default:
		return store.ArticleRead{}, err
	}
}

// GetArticleReadStatus reports whether the user has read the article (by slug).
// found is false when there is no read record (the article exists but is unread).
func (s *Service) GetArticleReadStatus(ctx context.Context, userID, slug string) (store.ArticleRead, bool, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return store.ArticleRead{}, false, apierr.Unauthorized("invalid user")
	}
	article, err := s.resolveArticle(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.ArticleRead{}, false, apierr.NotFound("article not found")
	}
	if err != nil {
		return store.ArticleRead{}, false, err
	}
	read, err := s.queries.GetArticleRead(ctx, store.GetArticleReadParams{UserID: uid, ArticleID: article.ID})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.ArticleRead{}, false, nil
	}
	if err != nil {
		return store.ArticleRead{}, false, err
	}
	return read, true, nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
