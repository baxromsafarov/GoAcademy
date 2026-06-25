package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// TrackItemInput is one entry in a track's ordered program.
type TrackItemInput struct {
	ContentType string
	ContentID   string
}

// TrackInput is the create/update payload for a track and its items.
type TrackInput struct {
	Title       string
	Description string
	Level       string
	Position    int
	Language    string
	Items       []TrackItemInput
}

// validate checks the input and returns the parsed item content IDs.
func (in TrackInput) validate() ([]pgtype.UUID, error) {
	details := map[string]string{}
	if strings.TrimSpace(in.Title) == "" {
		details["title"] = "must not be empty"
	}
	if !difficulties[in.Level] {
		details["level"] = "must be beginner, intermediate or advanced"
	}
	if !locales[in.Language] {
		details["language"] = "must be ru, en, uz or ja"
	}
	if in.Position < 0 {
		details["position"] = "must be >= 0"
	}
	ids := make([]pgtype.UUID, len(in.Items))
	for i, it := range in.Items {
		key := fmt.Sprintf("items[%d]", i)
		if !trackContentTypes[it.ContentType] {
			details[key+".content_type"] = "must be video, article, quiz, problem or project"
		}
		cid, err := pgxutil.ParseUUID(it.ContentID)
		if err != nil {
			details[key+".content_id"] = "must be a valid uuid"
		} else {
			ids[i] = cid
		}
	}
	if len(details) > 0 {
		return nil, apierr.Validation("invalid track").WithDetails(details)
	}
	return ids, nil
}

// CreateTrack inserts a track and its ordered items atomically.
func (s *Service) CreateTrack(ctx context.Context, in TrackInput) (store.Track, error) {
	ids, err := in.validate()
	if err != nil {
		return store.Track{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.Track{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	track, err := q.CreateTrack(ctx, store.CreateTrackParams{
		Title: in.Title, Description: in.Description, Level: store.Difficulty(in.Level),
		Position: int32(in.Position), Language: store.Locale(in.Language),
	})
	if err != nil {
		return store.Track{}, err
	}
	if err := insertTrackItems(ctx, q, track.ID, in.Items, ids); err != nil {
		return store.Track{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.Track{}, err
	}
	return track, nil
}

// UpdateTrack replaces a track's metadata and its items atomically.
func (s *Service) UpdateTrack(ctx context.Context, id string, in TrackInput) (store.Track, error) {
	tid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Track{}, apierr.NotFound("track not found")
	}
	ids, err := in.validate()
	if err != nil {
		return store.Track{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.Track{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	track, err := q.UpdateTrack(ctx, store.UpdateTrackParams{
		ID: tid, Title: in.Title, Description: in.Description, Level: store.Difficulty(in.Level),
		Position: int32(in.Position), Language: store.Locale(in.Language),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Track{}, apierr.NotFound("track not found")
	}
	if err != nil {
		return store.Track{}, err
	}
	if err := q.DeleteTrackItems(ctx, tid); err != nil {
		return store.Track{}, err
	}
	if err := insertTrackItems(ctx, q, tid, in.Items, ids); err != nil {
		return store.Track{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.Track{}, err
	}
	return track, nil
}

func insertTrackItems(ctx context.Context, q *store.Queries, trackID pgtype.UUID, items []TrackItemInput, ids []pgtype.UUID) error {
	for i, it := range items {
		if _, err := q.CreateTrackItem(ctx, store.CreateTrackItemParams{
			TrackID: trackID, ContentType: store.TrackContentType(it.ContentType),
			ContentID: ids[i], Position: int32(i + 1),
		}); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTrack removes a track (cascading its items).
func (s *Service) DeleteTrack(ctx context.Context, id string) error {
	tid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("track not found")
	}
	n, err := s.queries.DeleteTrack(ctx, tid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("track not found")
	}
	return nil
}
