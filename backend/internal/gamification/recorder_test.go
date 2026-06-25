package gamification

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run gamification integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func seedUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	email := fmt.Sprintf("gam-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(context.Background(), store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "G", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	return pgxutil.UUIDString(u.ID)
}

func newUUIDString(t *testing.T) string {
	t.Helper()
	u, err := pgxutil.NewUUID()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	return pgxutil.UUIDString(u)
}

func TestRecorder_AwardsXPAndIsIdempotent_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := NewRecorder(pool)
	svc := NewService(pool)
	uid := seedUser(t, pool)

	// A fresh user has no stats row yet: level 1, zero XP.
	if st, err := svc.GetStats(ctx, uid); err != nil || st.TotalXP != 0 || st.Level != 1 || st.LastActiveDate != "" {
		t.Fatalf("fresh stats = %+v, err %v; want level 1 / 0 xp / no date", st, err)
	}

	ref := newUUIDString(t)
	ev := activity.Event{UserID: uid, Type: "video_completed", RefType: "video", RefID: ref}
	if err := rec.Record(ctx, ev); err != nil {
		t.Fatalf("record: %v", err)
	}

	st, err := svc.GetStats(ctx, uid)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if st.TotalXP != 10 || st.Level != 1 {
		t.Errorf("after one video_completed: xp=%d level=%d, want 10/1", st.TotalXP, st.Level)
	}
	if st.LastActiveDate == "" {
		t.Error("last_active_date should be set after activity")
	}

	// Re-recording the SAME (user, type, ref): activity is logged again (heatmap)
	// but no extra XP is awarded.
	if err := rec.Record(ctx, ev); err != nil {
		t.Fatalf("record repeat: %v", err)
	}
	st2, _ := svc.GetStats(ctx, uid)
	if st2.TotalXP != 10 {
		t.Errorf("repeat must not re-award XP, got %d", st2.TotalXP)
	}
	cnt, err := store.New(pool).CountUserActivityByType(ctx, store.CountUserActivityByTypeParams{
		UserID: mustUUID(t, uid), ActivityType: "video_completed",
	})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if cnt != 2 {
		t.Errorf("activity should be journaled twice, got %d", cnt)
	}

	// A different ref earns XP again.
	if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "video_completed", RefType: "video", RefID: newUUIDString(t)}); err != nil {
		t.Fatalf("record different ref: %v", err)
	}
	st3, _ := svc.GetStats(ctx, uid)
	if st3.TotalXP != 20 {
		t.Errorf("distinct content should add XP, got %d", st3.TotalXP)
	}
}

func TestRecorder_LevelUp_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := NewRecorder(pool)
	svc := NewService(pool)
	uid := seedUser(t, pool)

	// Four distinct solved problems = 120 XP -> level 1 + floor(sqrt(1.2)) = 2.
	for i := 0; i < 4; i++ {
		if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "problem_solved", RefType: "problem", RefID: newUUIDString(t)}); err != nil {
			t.Fatalf("record solve %d: %v", i, err)
		}
	}
	st, _ := svc.GetStats(ctx, uid)
	if st.TotalXP != 120 || st.Level != 2 {
		t.Errorf("after 4 solves: xp=%d level=%d, want 120/2", st.TotalXP, st.Level)
	}
}

func TestRecorder_ConcurrentNoLostUpdates_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := NewRecorder(pool)
	svc := NewService(pool)
	uid := seedUser(t, pool)

	const n = 20 // each a distinct solved problem worth 30 XP
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		ref := newUUIDString(t)
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- rec.Record(ctx, activity.Event{UserID: uid, Type: "problem_solved", RefType: "problem", RefID: ref})
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent record: %v", err)
		}
	}

	st, _ := svc.GetStats(ctx, uid)
	if want := n * 30; st.TotalXP != want {
		t.Errorf("concurrent awards lost updates: xp=%d, want %d", st.TotalXP, want)
	}
	if want := LevelForXP(n * 30); st.Level != want {
		t.Errorf("level=%d, want %d", st.Level, want)
	}
}

func mustUUID(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	u, err := pgxutil.ParseUUID(s)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return u
}
