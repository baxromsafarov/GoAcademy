package progress

import (
	"context"
	"errors"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// ItemProgress is the completion state of one track item.
type ItemProgress struct {
	ContentType string
	ContentID   string
	Position    int32
	Completed   bool
}

// TrackProgressResult aggregates a user's completion of a track.
type TrackProgressResult struct {
	TrackID       string
	Total         int
	Completed     int
	Percent       int
	TrackComplete bool
	Items         []ItemProgress
}

// summarize computes totals, percent and completion from per-item flags. An empty
// track is 0% and not complete.
func summarize(flags []bool) (total, done, percent int, complete bool) {
	total = len(flags)
	for _, f := range flags {
		if f {
			done++
		}
	}
	if total > 0 {
		percent = int(math.Round(float64(done) * 100 / float64(total)))
		complete = done == total
	}
	return total, done, percent, complete
}

// TrackProgress computes the user's progress over a track. An item counts as
// completed when its underlying content is done: a watched video, a read article,
// a passed quiz, or a solved problem. (Projects are completable from CHAPTER 9.)
func (s *Service) TrackProgress(ctx context.Context, userID, trackID string) (TrackProgressResult, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return TrackProgressResult{}, apierr.Unauthorized("invalid user")
	}
	tid, err := pgxutil.ParseUUID(trackID)
	if err != nil {
		return TrackProgressResult{}, apierr.NotFound("track not found")
	}
	if _, err := s.queries.GetTrackByID(ctx, tid); errors.Is(err, pgx.ErrNoRows) {
		return TrackProgressResult{}, apierr.NotFound("track not found")
	} else if err != nil {
		return TrackProgressResult{}, err
	}

	items, err := s.queries.ListTrackItems(ctx, tid)
	if err != nil {
		return TrackProgressResult{}, err
	}

	var videoIDs, articleIDs, quizIDs, problemIDs []pgtype.UUID
	for _, it := range items {
		switch it.ContentType {
		case store.TrackContentTypeVideo:
			videoIDs = append(videoIDs, it.ContentID)
		case store.TrackContentTypeArticle:
			articleIDs = append(articleIDs, it.ContentID)
		case store.TrackContentTypeQuiz:
			quizIDs = append(quizIDs, it.ContentID)
		case store.TrackContentTypeProblem:
			problemIDs = append(problemIDs, it.ContentID)
		}
	}

	completed := make(map[string]struct{})
	add := func(typ string, ids []pgtype.UUID) {
		for _, id := range ids {
			completed[typ+":"+pgxutil.UUIDString(id)] = struct{}{}
		}
	}

	if len(videoIDs) > 0 {
		ids, err := s.queries.CompletedVideoIDs(ctx, store.CompletedVideoIDsParams{UserID: uid, Ids: videoIDs})
		if err != nil {
			return TrackProgressResult{}, err
		}
		add("video", ids)
	}
	if len(articleIDs) > 0 {
		ids, err := s.queries.ReadArticleIDs(ctx, store.ReadArticleIDsParams{UserID: uid, Ids: articleIDs})
		if err != nil {
			return TrackProgressResult{}, err
		}
		add("article", ids)
	}
	if len(quizIDs) > 0 {
		ids, err := s.queries.PassedQuizIDs(ctx, store.PassedQuizIDsParams{UserID: uid, Ids: quizIDs})
		if err != nil {
			return TrackProgressResult{}, err
		}
		add("quiz", ids)
	}
	if len(problemIDs) > 0 {
		ids, err := s.queries.SolvedProblemIDs(ctx, store.SolvedProblemIDsParams{UserID: uid, Ids: problemIDs})
		if err != nil {
			return TrackProgressResult{}, err
		}
		add("problem", ids)
	}

	flags := make([]bool, len(items))
	itemProgress := make([]ItemProgress, len(items))
	for i, it := range items {
		cid := pgxutil.UUIDString(it.ContentID)
		_, done := completed[string(it.ContentType)+":"+cid]
		flags[i] = done
		itemProgress[i] = ItemProgress{
			ContentType: string(it.ContentType),
			ContentID:   cid,
			Position:    it.Position,
			Completed:   done,
		}
	}

	total, done, percent, complete := summarize(flags)
	return TrackProgressResult{
		TrackID:       pgxutil.UUIDString(tid),
		Total:         total,
		Completed:     done,
		Percent:       percent,
		TrackComplete: complete,
		Items:         itemProgress,
	}, nil
}
