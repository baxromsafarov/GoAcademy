package activity

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// DBRecorder persists events to the activity_log table — the single source of
// truth for the activity heatmap, the period leaderboard and XP.
type DBRecorder struct {
	queries *store.Queries
}

// NewDBRecorder builds a DBRecorder over any store.DBTX (a connection pool or a
// transaction). Passing a pgx.Tx lets a caller record activity inside its own
// transaction, which CHAPTER 11 relies on to award XP atomically with the action.
func NewDBRecorder(db store.DBTX) *DBRecorder {
	return &DBRecorder{queries: store.New(db)}
}

// Record inserts exactly one activity_log row. RefID may be empty for ref-less
// events (stored as NULL); an invalid UserID or RefID is reported without
// touching the database.
func (r *DBRecorder) Record(ctx context.Context, ev Event) error {
	uid, err := pgxutil.ParseUUID(ev.UserID)
	if err != nil {
		return fmt.Errorf("activity: invalid user id %q: %w", ev.UserID, err)
	}

	var refID pgtype.UUID // zero value is NULL (Valid=false)
	if ev.RefID != "" {
		refID, err = pgxutil.ParseUUID(ev.RefID)
		if err != nil {
			return fmt.Errorf("activity: invalid ref id %q: %w", ev.RefID, err)
		}
	}

	_, err = r.queries.InsertActivity(ctx, store.InsertActivityParams{
		UserID:       uid,
		ActivityType: ev.Type,
		RefType:      ev.RefType,
		RefID:        refID,
		XpEarned:     int32(ev.XP),
	})
	return err
}
