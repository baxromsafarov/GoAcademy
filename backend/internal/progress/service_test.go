package progress

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func TestVideoProgressInput_Validate(t *testing.T) {
	if err := (VideoProgressInput{Percent: 150}).validate(); err == nil {
		t.Error("percent > 100 should be rejected")
	}
	if err := (VideoProgressInput{Percent: -1}).validate(); err == nil {
		t.Error("negative percent should be rejected")
	}
	if err := (VideoProgressInput{Position: -1}).validate(); err == nil {
		t.Error("negative position should be rejected")
	}
	if err := (VideoProgressInput{Percent: 50, Position: 10}).validate(); err != nil {
		t.Errorf("valid input should pass: %v", err)
	}
}

// countRecorder counts how many activity events were recorded.
type countRecorder struct{ n int }

func (c *countRecorder) Record(context.Context, activity.Event) error { c.n++; return nil }

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run progress integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestService_RecordVideoProgress_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := &countRecorder{}
	svc := NewService(pool, rec)

	// Seed a user (cascade-deletes its progress on cleanup) and a video.
	email := fmt.Sprintf("prog-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "P", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	marker := fmt.Sprintf("prog-%d", time.Now().UnixNano())
	var videoID string
	if err := pool.QueryRow(ctx,
		"INSERT INTO videos (title, youtube_id, tags) VALUES ('V', 'yt', ARRAY[$1]) RETURNING id::text", marker,
	).Scan(&videoID); err != nil {
		t.Fatalf("seed video: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM videos WHERE $1 = ANY(tags)", marker) })

	// 50% -> not completed.
	p, err := svc.RecordVideoProgress(ctx, userID, videoID, VideoProgressInput{Percent: 50, Position: 30})
	if err != nil {
		t.Fatalf("record 50%%: %v", err)
	}
	if p.Completed || p.WatchedPercent != 50 || p.LastPositionSeconds != 30 {
		t.Fatalf("unexpected after 50%%: %+v", p)
	}

	// 30% (lower) -> watched stays 50 (GREATEST), position updates.
	p2, _ := svc.RecordVideoProgress(ctx, userID, videoID, VideoProgressInput{Percent: 30, Position: 20})
	if p2.WatchedPercent != 50 {
		t.Errorf("watched_percent should not decrease, got %d", p2.WatchedPercent)
	}
	if p2.LastPositionSeconds != 20 {
		t.Errorf("position should update to 20, got %d", p2.LastPositionSeconds)
	}

	// 95% -> auto-completed, activity recorded once.
	p3, _ := svc.RecordVideoProgress(ctx, userID, videoID, VideoProgressInput{Percent: 95, Position: 100})
	if !p3.Completed {
		t.Error("95% should auto-complete")
	}
	if rec.n != 1 {
		t.Errorf("activity should be recorded once, got %d", rec.n)
	}

	// 10% after completion -> still completed, watched stays 95, no new activity.
	p4, _ := svc.RecordVideoProgress(ctx, userID, videoID, VideoProgressInput{Percent: 10, Position: 5})
	if !p4.Completed {
		t.Error("completed must be sticky")
	}
	if p4.WatchedPercent != 95 {
		t.Errorf("watched_percent should stay 95, got %d", p4.WatchedPercent)
	}
	if rec.n != 1 {
		t.Errorf("activity must not fire again on a repeat, got %d", rec.n)
	}

	// Manual mark on a fresh video at low percent -> completed.
	var videoID2 string
	if err := pool.QueryRow(ctx,
		"INSERT INTO videos (title, youtube_id, tags) VALUES ('V2', 'yt2', ARRAY[$1]) RETURNING id::text", marker,
	).Scan(&videoID2); err != nil {
		t.Fatalf("seed video2: %v", err)
	}
	manual := true
	pm, _ := svc.RecordVideoProgress(ctx, userID, videoID2, VideoProgressInput{Percent: 10, Position: 5, Completed: &manual})
	if !pm.Completed {
		t.Error("manual completed=true should complete regardless of percent")
	}

	// Unknown video -> 404 (foreign key).
	if _, err := svc.RecordVideoProgress(ctx, userID, "00000000-0000-0000-0000-000000000000", VideoProgressInput{Percent: 10}); err == nil {
		t.Error("unknown video should be not found")
	}
}

func TestService_MarkArticleRead_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := &countRecorder{}
	svc := NewService(pool, rec)

	email := fmt.Sprintf("aread-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "A", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	marker := fmt.Sprintf("aread-%d", time.Now().UnixNano())
	slug := "read-me-" + marker
	if _, err := pool.Exec(ctx, "INSERT INTO articles (title, slug, tags) VALUES ('R', $1, ARRAY[$2])", slug, marker); err != nil {
		t.Fatalf("seed article: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM articles WHERE $1 = ANY(tags)", marker) })

	// First read: recorded, activity fires once.
	r1, err := svc.MarkArticleRead(ctx, userID, slug)
	if err != nil {
		t.Fatalf("MarkArticleRead: %v", err)
	}
	if !r1.CompletedAt.Valid {
		t.Error("completed_at should be set")
	}
	if rec.n != 1 {
		t.Errorf("activity should fire once, got %d", rec.n)
	}

	// Repeat: idempotent, same completed_at, no new activity.
	r2, err := svc.MarkArticleRead(ctx, userID, slug)
	if err != nil {
		t.Fatalf("MarkArticleRead (repeat): %v", err)
	}
	if !r2.CompletedAt.Time.Equal(r1.CompletedAt.Time) {
		t.Error("completed_at should be stable on a repeat read")
	}
	if rec.n != 1 {
		t.Errorf("activity must not fire again on a repeat, got %d", rec.n)
	}

	// Unknown slug -> not found.
	if _, err := svc.MarkArticleRead(ctx, userID, "no-such-slug-here"); err == nil {
		t.Error("unknown slug should be not found")
	}
}

func TestService_ProblemSubmission_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := &countRecorder{}
	svc := NewService(pool, rec)

	email := fmt.Sprintf("psub-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "P", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	marker := fmt.Sprintf("psub-%d", time.Now().UnixNano())
	slug := "prob-" + marker
	if _, err := pool.Exec(ctx,
		"INSERT INTO problems (title, slug, reference_solution_markdown, tags) VALUES ('P', $1, 'THE-SOLUTION', ARRAY[$2])", slug, marker,
	); err != nil {
		t.Fatalf("seed problem: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM problems WHERE $1 = ANY(tags)", marker) })

	// Attempted: no solution revealed, solution endpoint forbidden, no activity.
	r1, err := svc.SubmitProblem(ctx, userID, slug, SubmitProblemInput{Code: "package main", Language: "go", Solved: false})
	if err != nil {
		t.Fatalf("SubmitProblem (attempted): %v", err)
	}
	if r1.Submission.Status != store.SubmissionStatusAttempted {
		t.Errorf("status = %q, want attempted", r1.Submission.Status)
	}
	if r1.ReferenceSolution != "" {
		t.Error("attempted submission must not reveal the solution")
	}
	if _, err := svc.GetProblemSolution(ctx, userID, slug); err == nil {
		t.Error("solution must be forbidden before solving")
	}
	if rec.n != 0 {
		t.Errorf("no activity expected yet, got %d", rec.n)
	}

	// Solved: solution revealed, activity fires once.
	r2, err := svc.SubmitProblem(ctx, userID, slug, SubmitProblemInput{Code: "solved code", Solved: true})
	if err != nil {
		t.Fatalf("SubmitProblem (solved): %v", err)
	}
	if r2.Submission.Status != store.SubmissionStatusSolved {
		t.Errorf("status = %q, want solved", r2.Submission.Status)
	}
	if r2.ReferenceSolution != "THE-SOLUTION" {
		t.Errorf("solved submission should reveal the solution, got %q", r2.ReferenceSolution)
	}
	if rec.n != 1 {
		t.Errorf("activity should fire once, got %d", rec.n)
	}

	// Solution endpoint now returns it.
	sol, err := svc.GetProblemSolution(ctx, userID, slug)
	if err != nil || sol != "THE-SOLUTION" {
		t.Errorf("GetProblemSolution = (%q, %v), want THE-SOLUTION", sol, err)
	}

	// Submitting solved again: solution still revealed, activity not re-fired.
	r3, _ := svc.SubmitProblem(ctx, userID, slug, SubmitProblemInput{Solved: true})
	if r3.ReferenceSolution != "THE-SOLUTION" {
		t.Error("solution should remain revealed for an already-solved problem")
	}
	if rec.n != 1 {
		t.Errorf("activity must not re-fire, got %d", rec.n)
	}

	// Unknown problem -> not found.
	if _, err := svc.SubmitProblem(ctx, userID, "no-such-problem", SubmitProblemInput{}); err == nil {
		t.Error("unknown problem should be not found")
	}
}

func scanID(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&id); err != nil {
		t.Fatalf("seed (%s): %v", sql, err)
	}
	return id
}

func TestService_TrackProgress_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool, &countRecorder{})

	email := fmt.Sprintf("trk-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "T", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	marker := fmt.Sprintf("trkprog-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM videos WHERE $1 = ANY(tags)", marker)
		_, _ = pool.Exec(context.Background(), "DELETE FROM articles WHERE $1 = ANY(tags)", marker)
		_, _ = pool.Exec(context.Background(), "DELETE FROM quizzes WHERE $1 = ANY(tags)", marker)
		_, _ = pool.Exec(context.Background(), "DELETE FROM problems WHERE $1 = ANY(tags)", marker)
	})

	vid := scanID(t, pool, "INSERT INTO videos (title, youtube_id, tags) VALUES ('V','y',ARRAY[$1]) RETURNING id::text", marker)
	art := scanID(t, pool, "INSERT INTO articles (title, slug, tags) VALUES ('A',$1,ARRAY[$2]) RETURNING id::text", "a-"+marker, marker)
	qz := scanID(t, pool, "INSERT INTO quizzes (title, tags) VALUES ('Q',ARRAY[$1]) RETURNING id::text", marker)
	prob := scanID(t, pool, "INSERT INTO problems (title, slug, tags) VALUES ('P',$1,ARRAY[$2]) RETURNING id::text", "p-"+marker, marker)

	track := scanID(t, pool, "INSERT INTO tracks (title, language) VALUES ($1,'en') RETURNING id::text", "T-"+marker)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM tracks WHERE id = $1", track) })
	if _, err := pool.Exec(ctx,
		`INSERT INTO track_items (track_id, content_type, content_id, position) VALUES
		 ($1,'video',$2::uuid,1),($1,'article',$3::uuid,2),($1,'quiz',$4::uuid,3),($1,'problem',$5::uuid,4)`,
		track, vid, art, qz, prob); err != nil {
		t.Fatalf("seed track items: %v", err)
	}

	// Nothing completed yet.
	r0, err := svc.TrackProgress(ctx, userID, track)
	if err != nil {
		t.Fatalf("TrackProgress: %v", err)
	}
	if r0.Total != 4 || r0.Completed != 0 || r0.Percent != 0 || r0.TrackComplete {
		t.Errorf("initial: total=%d done=%d pct=%d complete=%v, want 4/0/0/false", r0.Total, r0.Completed, r0.Percent, r0.TrackComplete)
	}

	// Complete the video and the article.
	if _, err := pool.Exec(ctx, "INSERT INTO video_progress (user_id, video_id, watched_percent, completed) VALUES ($1::uuid,$2::uuid,100,true)", userID, vid); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO article_reads (user_id, article_id) VALUES ($1::uuid,$2::uuid)", userID, art); err != nil {
		t.Fatal(err)
	}
	r1, _ := svc.TrackProgress(ctx, userID, track)
	if r1.Completed != 2 || r1.Percent != 50 || r1.TrackComplete {
		t.Errorf("half: done=%d pct=%d complete=%v, want 2/50/false", r1.Completed, r1.Percent, r1.TrackComplete)
	}

	// Pass the quiz and solve the problem.
	if _, err := pool.Exec(ctx, "INSERT INTO quiz_attempts (user_id, quiz_id, score, passed) VALUES ($1::uuid,$2::uuid,100,true)", userID, qz); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO problem_submissions (user_id, problem_id, status) VALUES ($1::uuid,$2::uuid,'solved')", userID, prob); err != nil {
		t.Fatal(err)
	}
	r2, _ := svc.TrackProgress(ctx, userID, track)
	if r2.Completed != 4 || r2.Percent != 100 || !r2.TrackComplete {
		t.Errorf("full: done=%d pct=%d complete=%v, want 4/100/true", r2.Completed, r2.Percent, r2.TrackComplete)
	}

	if _, err := svc.TrackProgress(ctx, userID, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown track should be not found")
	}
}

func TestService_ProjectChecklist_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := &countRecorder{}
	svc := NewService(pool, rec)

	email := fmt.Sprintf("proj-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "PR", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	marker := fmt.Sprintf("proj-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM mini_projects WHERE $1 = ANY(tags)", marker)
	})
	pid := scanID(t, pool, "INSERT INTO mini_projects (title, tags) VALUES ('P', ARRAY[$1]) RETURNING id::text", marker)
	s1 := scanID(t, pool, "INSERT INTO mini_project_steps (project_id, text, position) VALUES ($1,'s1',1) RETURNING id::text", pid)
	s2 := scanID(t, pool, "INSERT INTO mini_project_steps (project_id, text, position) VALUES ($1,'s2',2) RETURNING id::text", pid)
	s3 := scanID(t, pool, "INSERT INTO mini_project_steps (project_id, text, position) VALUES ($1,'s3',3) RETURNING id::text", pid)

	// Initial: nothing done.
	p0, err := svc.ProjectProgress(ctx, userID, pid)
	if err != nil {
		t.Fatalf("ProjectProgress: %v", err)
	}
	if p0.Total != 3 || p0.Completed != 0 || p0.ProjectComplete {
		t.Errorf("initial: total=%d done=%d complete=%v, want 3/0/false", p0.Total, p0.Completed, p0.ProjectComplete)
	}

	// Toggle s1 on.
	r1, _ := svc.ToggleProjectStep(ctx, userID, pid, s1)
	if r1.Completed != 1 || len(r1.CompletedStepIDs) != 1 || r1.CompletedStepIDs[0] != s1 {
		t.Errorf("after toggle s1 on: %+v", r1)
	}
	if rec.n != 0 {
		t.Errorf("no activity expected yet, got %d", rec.n)
	}

	// Toggle s1 off (pressing again undoes it).
	r1off, _ := svc.ToggleProjectStep(ctx, userID, pid, s1)
	if r1off.Completed != 0 {
		t.Errorf("toggle should be reversible, got %d completed", r1off.Completed)
	}

	// Complete all three -> project complete, activity fires once.
	_, _ = svc.ToggleProjectStep(ctx, userID, pid, s1)
	_, _ = svc.ToggleProjectStep(ctx, userID, pid, s2)
	rall, _ := svc.ToggleProjectStep(ctx, userID, pid, s3)
	if rall.Completed != 3 || !rall.ProjectComplete {
		t.Errorf("after all: done=%d complete=%v, want 3/true", rall.Completed, rall.ProjectComplete)
	}
	if rec.n != 1 {
		t.Errorf("project_completed activity should fire once, got %d", rec.n)
	}

	// Invalid step / unknown project.
	if _, err := svc.ToggleProjectStep(ctx, userID, pid, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown step should be not found")
	}
	if _, err := svc.ToggleProjectStep(ctx, userID, "00000000-0000-0000-0000-000000000000", s1); err == nil {
		t.Error("unknown project should be not found")
	}
}
