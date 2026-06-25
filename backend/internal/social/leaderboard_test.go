package social

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func TestPeriodWindow(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	day := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

	for _, p := range []string{"all", ""} {
		if _, _, allTime, err := periodWindow(p, now); err != nil || !allTime {
			t.Errorf("periodWindow(%q): allTime=%v err=%v, want allTime=true", p, allTime, err)
		}
	}
	if from, to, allTime, err := periodWindow("week", now); err != nil || allTime ||
		!from.Equal(day(2026, 6, 19)) || !to.Equal(day(2026, 6, 26)) {
		t.Errorf("week = [%v,%v) allTime=%v err=%v, want [2026-06-19,2026-06-26)", from, to, allTime, err)
	}
	if from, to, _, err := periodWindow("month", now); err != nil ||
		!from.Equal(day(2026, 5, 27)) || !to.Equal(day(2026, 6, 26)) {
		t.Errorf("month = [%v,%v) err=%v, want [2026-05-27,2026-06-26)", from, to, err)
	}
	if _, _, _, err := periodWindow("year", now); err == nil {
		t.Error("invalid period should error")
	}
}

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run social integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedUserStats creates a user with the given visibility and a user_stats row.
func seedUserStats(t *testing.T, pool *pgxpool.Pool, tag string, isPublic, isBlocked bool, xp int) string {
	t.Helper()
	ctx := context.Background()
	email := fmt.Sprintf("%s-%d@example.com", tag, time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: tag, Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user %s: %v", tag, err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	uid := pgxutil.UUIDString(u.ID)
	if _, err := pool.Exec(ctx, "UPDATE users SET is_public = $2, is_blocked = $3 WHERE id = $1::uuid", uid, isPublic, isBlocked); err != nil {
		t.Fatalf("set visibility: %v", err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO user_stats (user_id, total_xp, level) VALUES ($1::uuid, $2, 1)", uid, xp); err != nil {
		t.Fatalf("seed stats: %v", err)
	}
	return uid
}

func TestLeaderboard_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	now := time.Now().UTC()

	// High XP keeps the seeded users at the top regardless of other test data.
	a := seedUserStats(t, pool, "lb-a", true, false, 1_000_000)  // public
	b := seedUserStats(t, pool, "lb-b", true, false, 999_999)    // public, lower
	c := seedUserStats(t, pool, "lb-c", false, false, 1_000_001) // private -> hidden
	d := seedUserStats(t, pool, "lb-d", true, true, 1_000_002)   // blocked -> hidden

	index := func(entries []LeaderboardEntry) map[string]LeaderboardEntry {
		m := map[string]LeaderboardEntry{}
		for _, e := range entries {
			m[e.UserID] = e
		}
		return m
	}

	// All-time ranking.
	all, err := svc.Leaderboard(ctx, "all", now, 100, 0)
	if err != nil {
		t.Fatalf("Leaderboard all: %v", err)
	}
	m := index(all)
	if m[a].XP != 1_000_000 || m[b].XP != 999_999 {
		t.Errorf("all-time XP wrong: a=%d b=%d", m[a].XP, m[b].XP)
	}
	if _, ok := m[c]; ok {
		t.Error("private user must be hidden")
	}
	if _, ok := m[d]; ok {
		t.Error("blocked user must be hidden")
	}
	if m[a].Rank >= m[b].Rank {
		t.Errorf("A (more XP) should rank above B: a=%d b=%d", m[a].Rank, m[b].Rank)
	}

	// Period ranking: A active in-window, B active out-of-window.
	mustExec(t, pool, "INSERT INTO activity_log (user_id, activity_type, ref_type, xp_earned, occurred_at) VALUES ($1::uuid,'x','',$2,$3)", a, 40, now)
	mustExec(t, pool, "INSERT INTO activity_log (user_id, activity_type, ref_type, xp_earned, occurred_at) VALUES ($1::uuid,'x','',$2,$3)", b, 10, now.AddDate(0, 0, -60))

	week, err := svc.Leaderboard(ctx, "week", now, 100, 0)
	if err != nil {
		t.Fatalf("Leaderboard week: %v", err)
	}
	wm := index(week)
	if wm[a].XP != 40 {
		t.Errorf("period XP for A should be 40 (window sum), got %d", wm[a].XP)
	}
	if _, ok := wm[b]; ok {
		t.Error("B has no in-window activity and must be absent from the period board")
	}

	// Invalid period.
	if _, err := svc.Leaderboard(ctx, "decade", now, 100, 0); err == nil {
		t.Error("invalid period should error")
	}
}

func mustExec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec: %v", err)
	}
}
