package gamification

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// DailyChallenge is the challenge offered on a given day, plus whether the
// requesting user has completed it.
type DailyChallenge struct {
	Date        string // "YYYY-MM-DD" (UTC)
	ContentType string
	ContentID   string
	BonusXP     int
	Completed   bool
}

// DailyCompletion is the outcome of completing today's challenge.
type DailyCompletion struct {
	Challenge      DailyChallenge
	NewlyCompleted bool
}

// DailyService serves the daily challenge and records its completion.
type DailyService struct {
	queries  *store.Queries
	recorder *Recorder
}

// NewDailyService wires the daily-challenge service; the recorder awards the
// bonus XP (and streak/badges) on first completion.
func NewDailyService(pool *pgxpool.Pool, recorder *Recorder) *DailyService {
	return &DailyService{queries: store.New(pool), recorder: recorder}
}

// Today returns the challenge for the given UTC day and whether the user has done
// it. Returns a 404 domain error when there is no challenge that day.
func (s *DailyService) Today(ctx context.Context, userID string, day time.Time) (DailyChallenge, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return DailyChallenge{}, apierr.Unauthorized("invalid user")
	}
	ch, err := s.queries.GetDailyChallengeByDate(ctx, pgtype.Date{Time: day, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return DailyChallenge{}, apierr.NotFound("no daily challenge today")
	}
	if err != nil {
		return DailyChallenge{}, err
	}
	done, err := s.queries.IsDailyChallengeCompleted(ctx, store.IsDailyChallengeCompletedParams{
		UserID: uid, ChallengeID: ch.ID,
	})
	if err != nil {
		return DailyChallenge{}, err
	}
	return toDailyChallenge(ch, done), nil
}

// Complete marks today's challenge done for the user. It is idempotent: a repeat
// awards nothing further. On the first completion it records a
// "daily_challenge_completed" activity worth the challenge's bonus XP, which the
// recorder folds into XP, streak and badges.
func (s *DailyService) Complete(ctx context.Context, userID string, day time.Time) (DailyCompletion, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return DailyCompletion{}, apierr.Unauthorized("invalid user")
	}
	ch, err := s.queries.GetDailyChallengeByDate(ctx, pgtype.Date{Time: day, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return DailyCompletion{}, apierr.NotFound("no daily challenge today")
	}
	if err != nil {
		return DailyCompletion{}, err
	}

	_, err = s.queries.CompleteDailyChallenge(ctx, store.CompleteDailyChallengeParams{
		UserID: uid, ChallengeID: ch.ID,
	})
	switch {
	case err == nil:
		// First completion: award the bonus (best-effort, like other activity hooks).
		_ = s.recorder.Record(ctx, activity.Event{
			UserID:  userID,
			Type:    "daily_challenge_completed",
			RefType: "daily_challenge",
			RefID:   pgxutil.UUIDString(ch.ID),
			XP:      int(ch.BonusXp),
		})
		return DailyCompletion{Challenge: toDailyChallenge(ch, true), NewlyCompleted: true}, nil
	case errors.Is(err, pgx.ErrNoRows):
		return DailyCompletion{Challenge: toDailyChallenge(ch, true), NewlyCompleted: false}, nil
	default:
		return DailyCompletion{}, err
	}
}

func toDailyChallenge(ch store.DailyChallenge, completed bool) DailyChallenge {
	return DailyChallenge{
		Date:        ch.ChallengeDate.Time.Format("2006-01-02"),
		ContentType: string(ch.ContentType),
		ContentID:   pgxutil.UUIDString(ch.ContentID),
		BonusXP:     int(ch.BonusXp),
		Completed:   completed,
	}
}
