package gamification

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// Recorder implements activity.Recorder. It writes the activity_log row and the
// user_stats XP/level update in one transaction, so XP is always consistent with
// the activity that earned it.
type Recorder struct {
	pool *pgxpool.Pool
}

// NewRecorder builds a gamification-aware activity recorder.
func NewRecorder(pool *pgxpool.Pool) *Recorder { return &Recorder{pool: pool} }

// Record journals the event and updates the user's stats atomically. XP is granted
// only the first time a given (user, activity_type, ref) occurs — repeats are still
// logged (for the heatmap) but earn nothing, which prevents XP farming. The level
// is recomputed from the new lifetime total and the streak from the activity day;
// the user_stats row is locked for the transaction so concurrent updates serialize
// (no lost updates).
func (r *Recorder) Record(ctx context.Context, ev activity.Event) error {
	uid, err := pgxutil.ParseUUID(ev.UserID)
	if err != nil {
		return fmt.Errorf("gamification: invalid user id %q: %w", ev.UserID, err)
	}
	var refID pgtype.UUID // zero value is NULL (Valid=false)
	if ev.RefID != "" {
		refID, err = pgxutil.ParseUUID(ev.RefID)
		if err != nil {
			return fmt.Errorf("gamification: invalid ref id %q: %w", ev.RefID, err)
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := store.New(tx)

	// An explicit Event.XP (e.g. a daily challenge's variable bonus) overrides the
	// per-type default; otherwise XP is the type's standard award.
	xp := ev.XP
	if xp == 0 {
		xp = XPFor(ev.Type)
	}
	if xp > 0 && refID.Valid {
		seen, err := q.ActivityExists(ctx, store.ActivityExistsParams{
			UserID: uid, ActivityType: ev.Type, RefID: refID,
		})
		if err != nil {
			return err
		}
		if seen {
			xp = 0 // already earned for this content
		}
	}

	row, err := q.InsertActivity(ctx, store.InsertActivityParams{
		UserID:       uid,
		ActivityType: ev.Type,
		RefType:      ev.RefType,
		RefID:        refID,
		XpEarned:     int32(xp),
	})
	if err != nil {
		return err
	}
	newDate := utcDate(row.OccurredAt.Time)

	// Ensure the row exists, then lock it for the rest of the transaction so XP and
	// streak are computed and written without races.
	if err := q.EnsureUserStats(ctx, uid); err != nil {
		return err
	}
	cur, err := q.LockUserStats(ctx, uid)
	if err != nil {
		return err
	}

	newTotal := int(cur.TotalXp) + xp
	su := computeStreak(cur.LastActiveDate.Time, cur.LastActiveDate.Valid,
		int(cur.CurrentStreak), int(cur.LongestStreak), newDate)

	if err := q.UpdateUserStats(ctx, store.UpdateUserStatsParams{
		TotalXp:        int32(newTotal),
		Level:          int32(LevelForXP(newTotal)),
		CurrentStreak:  int32(su.Current),
		LongestStreak:  int32(su.Longest),
		LastActiveDate: pgtype.Date{Time: su.LastActive, Valid: true},
		UserID:         uid,
	}); err != nil {
		return err
	}

	// Award any newly-earned badges in the same transaction.
	if err := awardBadges(ctx, q, uid, newTotal, su.Current); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// utcDate truncates a timestamp to its UTC calendar day.
func utcDate(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
