package progress

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func d(y int, m time.Month, day int) time.Time {
	return time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
}

func TestParseHeatmapRange(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name             string
		from, to         string
		wantFrom, wantTo time.Time // wantTo is the exclusive end
		wantErr          bool
	}{
		{"both empty defaults to one-year window", "", "", d(2025, 6, 26), d(2026, 6, 26), false},
		{"explicit range", "2026-01-01", "2026-01-31", d(2026, 1, 1), d(2026, 2, 1), false},
		{"from only, to defaults to today", "2026-06-01", "", d(2026, 6, 1), d(2026, 6, 26), false},
		{"to only, from defaults a year back", "", "2026-03-10", d(2025, 3, 11), d(2026, 3, 11), false},
		{"single day", "2026-06-25", "2026-06-25", d(2026, 6, 25), d(2026, 6, 26), false},
		{"from after to", "2026-02-01", "2026-01-01", time.Time{}, time.Time{}, true},
		{"bad from format", "2026/06/01", "", time.Time{}, time.Time{}, true},
		{"bad to format", "", "not-a-date", time.Time{}, time.Time{}, true},
		{"span too large", "2020-01-01", "2026-01-01", time.Time{}, time.Time{}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			from, to, err := ParseHeatmapRange(tc.from, tc.to, now)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got from=%v to=%v", from, to)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !from.Equal(tc.wantFrom) || !to.Equal(tc.wantTo) {
				t.Errorf("got [%s, %s), want [%s, %s)",
					from.Format(dateLayout), to.Format(dateLayout),
					tc.wantFrom.Format(dateLayout), tc.wantTo.Format(dateLayout))
			}
		})
	}
}

func mustExec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec (%s): %v", sql, err)
	}
}

func TestService_ProgressSummary_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool, &countRecorder{})

	email := fmt.Sprintf("sum-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "S", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	// A fresh user has done nothing.
	s0, err := svc.ProgressSummary(ctx, userID)
	if err != nil {
		t.Fatalf("ProgressSummary: %v", err)
	}
	if s0 != (ProgressSummaryResult{}) {
		t.Fatalf("fresh user summary should be all zero, got %+v", s0)
	}

	marker := fmt.Sprintf("sum-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		for _, tbl := range []string{"videos", "articles", "quizzes", "problems", "mini_projects"} {
			_, _ = pool.Exec(context.Background(), "DELETE FROM "+tbl+" WHERE $1 = ANY(tags)", marker)
		}
	})

	vid := scanID(t, pool, "INSERT INTO videos (title, youtube_id, tags) VALUES ('V','y',ARRAY[$1]) RETURNING id::text", marker)
	art := scanID(t, pool, "INSERT INTO articles (title, slug, tags) VALUES ('A',$1,ARRAY[$2]) RETURNING id::text", "a-"+marker, marker)
	qz := scanID(t, pool, "INSERT INTO quizzes (title, tags) VALUES ('Q',ARRAY[$1]) RETURNING id::text", marker)
	prob := scanID(t, pool, "INSERT INTO problems (title, slug, tags) VALUES ('P',$1,ARRAY[$2]) RETURNING id::text", "p-"+marker, marker)
	pid := scanID(t, pool, "INSERT INTO mini_projects (title, tags) VALUES ('PR',ARRAY[$1]) RETURNING id::text", marker)
	st1 := scanID(t, pool, "INSERT INTO mini_project_steps (project_id, text, position) VALUES ($1,'s1',1) RETURNING id::text", pid)
	st2 := scanID(t, pool, "INSERT INTO mini_project_steps (project_id, text, position) VALUES ($1,'s2',2) RETURNING id::text", pid)

	// Complete exactly one of each section.
	mustExec(t, pool, "INSERT INTO video_progress (user_id, video_id, watched_percent, completed) VALUES ($1::uuid,$2::uuid,100,true)", userID, vid)
	mustExec(t, pool, "INSERT INTO article_reads (user_id, article_id) VALUES ($1::uuid,$2::uuid)", userID, art)
	mustExec(t, pool, "INSERT INTO quiz_attempts (user_id, quiz_id, score, passed) VALUES ($1::uuid,$2::uuid,100,true)", userID, qz)
	mustExec(t, pool, "INSERT INTO problem_submissions (user_id, problem_id, status) VALUES ($1::uuid,$2::uuid,'solved')", userID, prob)
	if _, err := svc.ToggleProjectStep(ctx, userID, pid, st1); err != nil {
		t.Fatalf("toggle s1: %v", err)
	}
	if _, err := svc.ToggleProjectStep(ctx, userID, pid, st2); err != nil {
		t.Fatalf("toggle s2: %v", err)
	}

	s1, err := svc.ProgressSummary(ctx, userID)
	if err != nil {
		t.Fatalf("ProgressSummary after seeding: %v", err)
	}
	want := ProgressSummaryResult{VideosCompleted: 1, ArticlesRead: 1, QuizzesPassed: 1, ProblemsSolved: 1, ProjectsCompleted: 1}
	if s1 != want {
		t.Errorf("summary = %+v, want %+v", s1, want)
	}
}

func TestService_ActivityHeatmap_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool, &countRecorder{})

	email := fmt.Sprintf("heat-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "H", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	// Two events on 2026-06-20 (UTC, incl. one near midnight), one on 06-22, and
	// one outside the queried window on 06-01.
	ins := func(ts time.Time, xp int) {
		mustExec(t, pool,
			"INSERT INTO activity_log (user_id, activity_type, ref_type, occurred_at, xp_earned) VALUES ($1::uuid,'x','',$2,$3)",
			userID, ts, xp)
	}
	ins(time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC), 5)
	ins(time.Date(2026, 6, 20, 23, 30, 0, 0, time.UTC), 3)
	ins(time.Date(2026, 6, 22, 5, 0, 0, 0, time.UTC), 10)
	ins(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 100) // out of range

	days, err := svc.ActivityHeatmap(ctx, userID, d(2026, 6, 15), d(2026, 6, 23))
	if err != nil {
		t.Fatalf("ActivityHeatmap: %v", err)
	}

	got := map[string]ActivityDay{}
	for _, dd := range days {
		got[dd.Day] = dd
	}
	if len(got) != 2 {
		t.Fatalf("want 2 active days in range, got %d: %+v", len(got), days)
	}
	if g := got["2026-06-20"]; g.Count != 2 || g.XP != 8 {
		t.Errorf("2026-06-20 = count %d xp %d, want 2/8", g.Count, g.XP)
	}
	if g := got["2026-06-22"]; g.Count != 1 || g.XP != 10 {
		t.Errorf("2026-06-22 = count %d xp %d, want 1/10", g.Count, g.XP)
	}
	if _, ok := got["2026-06-01"]; ok {
		t.Error("2026-06-01 is outside the window and must be excluded")
	}

	// Days are returned in ascending order.
	if len(days) == 2 && days[0].Day > days[1].Day {
		t.Errorf("days must be ordered ascending, got %s then %s", days[0].Day, days[1].Day)
	}
}
