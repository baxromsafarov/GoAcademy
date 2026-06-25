package activity_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/progress"
	"github.com/goacademy/backend/internal/store"
)

// failDBTX fails the test if any database method is called. It proves that
// DBRecorder rejects invalid input before issuing a query.
type failDBTX struct{ t *testing.T }

func (f failDBTX) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	f.t.Helper()
	f.t.Fatal("DB must not be touched for invalid input")
	return pgconn.CommandTag{}, nil
}

func (f failDBTX) Query(context.Context, string, ...any) (pgx.Rows, error) {
	f.t.Helper()
	f.t.Fatal("DB must not be touched for invalid input")
	return nil, nil
}

func (f failDBTX) QueryRow(context.Context, string, ...any) pgx.Row {
	f.t.Helper()
	f.t.Fatal("DB must not be touched for invalid input")
	return nil
}

func TestDBRecorder_InvalidIDs(t *testing.T) {
	ctx := context.Background()
	rec := activity.NewDBRecorder(failDBTX{t})

	if err := rec.Record(ctx, activity.Event{UserID: "not-a-uuid", Type: "x"}); err == nil {
		t.Error("invalid user id should error")
	}
	if err := rec.Record(ctx, activity.Event{
		UserID: "11111111-1111-1111-1111-111111111111", RefID: "nope", Type: "x",
	}); err == nil {
		t.Error("invalid ref id should error")
	}
}

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run activity integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedUser inserts a throwaway user and returns its UUID; deleting it on cleanup
// cascade-removes any activity_log rows it owns.
func seedUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	email := fmt.Sprintf("act-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(context.Background(), store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "A", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	return pgxutil.UUIDString(u.ID)
}

func TestDBRecorder_Persist_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	uid := seedUser(t, pool)
	rec := activity.NewDBRecorder(pool)
	q := store.New(pool)
	puid, _ := pgxutil.ParseUUID(uid)

	refID := "abcabcab-0000-4000-8000-000000000001" // any uuid; ref_id has no FK (D-006)
	if err := rec.Record(ctx, activity.Event{
		UserID: uid, Type: "video_completed", RefType: "video", RefID: refID, XP: 5,
	}); err != nil {
		t.Fatalf("record video_completed: %v", err)
	}
	// A ref-less event (no RefType/RefID) must store as empty/NULL.
	if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "daily_login"}); err != nil {
		t.Fatalf("record ref-less: %v", err)
	}

	rows, err := q.ListUserActivity(ctx, puid)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 activity rows, got %d", len(rows))
	}

	var vc, dl *store.ActivityLog
	for i := range rows {
		switch rows[i].ActivityType {
		case "video_completed":
			vc = &rows[i]
		case "daily_login":
			dl = &rows[i]
		}
	}
	if vc == nil || dl == nil {
		t.Fatalf("missing expected rows: %+v", rows)
	}
	if vc.RefType != "video" || !vc.RefID.Valid || pgxutil.UUIDString(vc.RefID) != refID ||
		vc.XpEarned != 5 || !vc.OccurredAt.Valid {
		t.Errorf("video_completed row persisted wrong: %+v", vc)
	}
	if dl.RefType != "" || dl.RefID.Valid {
		t.Errorf("ref-less row should have empty ref_type and NULL ref_id: %+v", dl)
	}

	n, err := q.CountUserActivityByType(ctx, store.CountUserActivityByTypeParams{
		UserID: puid, ActivityType: "video_completed",
	})
	if err != nil || n != 1 {
		t.Errorf("CountUserActivityByType = (%d, %v), want (1, nil)", n, err)
	}
}

// TestDBRecorder_NoDoubleRecords_Integration drives a real call site (video
// progress) with the real recorder and asserts the action is journaled exactly
// once, however many times progress is reported.
func TestDBRecorder_NoDoubleRecords_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	uid := seedUser(t, pool)
	svc := progress.NewService(pool, activity.NewDBRecorder(pool))

	marker := fmt.Sprintf("act-%d", time.Now().UnixNano())
	var videoID string
	if err := pool.QueryRow(ctx,
		"INSERT INTO videos (title, youtube_id, tags) VALUES ('V','yt',ARRAY[$1]) RETURNING id::text", marker,
	).Scan(&videoID); err != nil {
		t.Fatalf("seed video: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM videos WHERE $1 = ANY(tags)", marker) })

	for _, pct := range []int{50, 95, 10, 100} {
		if _, err := svc.RecordVideoProgress(ctx, uid, videoID, progress.VideoProgressInput{Percent: pct}); err != nil {
			t.Fatalf("record %d%%: %v", pct, err)
		}
	}

	puid, _ := pgxutil.ParseUUID(uid)
	n, err := store.New(pool).CountUserActivityByType(ctx, store.CountUserActivityByTypeParams{
		UserID: puid, ActivityType: "video_completed",
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Errorf("video_completed must be journaled exactly once, got %d", n)
	}
}
