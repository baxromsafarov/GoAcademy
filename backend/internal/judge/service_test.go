package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/runner"
	"github.com/goacademy/backend/internal/store"
)

type countRecorder struct{ n int }

func (c *countRecorder) Record(context.Context, activity.Event) error { c.n++; return nil }

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run judge integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestService_Submit_Integration(t *testing.T) {
	dockerReady(t)
	pool := openPool(t)
	ctx := context.Background()
	rec := &countRecorder{}
	svc := NewService(runner.New("busybox", ""), pool, rec, runner.Limits{WallTime: 3 * time.Second})

	// Seed a user, a problem, and its test cases (all cleaned up on exit).
	email := fmt.Sprintf("judge-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "J", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	slug := fmt.Sprintf("judge-test-%d", time.Now().UnixNano())
	var problemID string
	if err := pool.QueryRow(ctx,
		"INSERT INTO problems (title, slug) VALUES ('Sum', $1) RETURNING id::text", slug,
	).Scan(&problemID); err != nil {
		t.Fatalf("seed problem: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM problems WHERE slug = $1", slug) })

	for i, c := range sumCases() {
		if _, err := pool.Exec(ctx,
			"INSERT INTO problem_test_cases (problem_id, input, expected_output, is_sample, position) VALUES ($1,$2,$3,$4,$5)",
			problemID, c.Input, c.ExpectedOutput, c.IsSample, i,
		); err != nil {
			t.Fatalf("seed case: %v", err)
		}
	}

	// Correct submission → OK, persisted as solved, activity recorded once.
	sub, res, err := svc.Submit(ctx, userID, problemID, sumProgram, "go")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Verdict != OK {
		t.Fatalf("verdict = %s, want OK (%+v)", res.Verdict, res)
	}
	if sub.Status != store.SubmissionStatusSolved {
		t.Errorf("status = %s, want solved", sub.Status)
	}
	if rec.n != 1 {
		t.Errorf("activity recorded %d times, want 1", rec.n)
	}

	// The verdict JSON must be persisted on the submission row.
	var verdictJSON []byte
	if err := pool.QueryRow(ctx,
		"SELECT verdict FROM problem_submissions WHERE id::text = $1", pgxutil.UUIDString(sub.ID),
	).Scan(&verdictJSON); err != nil {
		t.Fatalf("read back verdict: %v", err)
	}
	var persisted Result
	if err := json.Unmarshal(verdictJSON, &persisted); err != nil {
		t.Fatalf("verdict json: %v", err)
	}
	if persisted.Verdict != OK || persisted.Passed != 3 || persisted.Total != 3 {
		t.Errorf("persisted verdict = %+v", persisted)
	}

	// Wrong submission → WA, persisted as attempted, no new solve activity.
	wrong := `package main
import ("bufio";"fmt";"os")
func main(){ r := bufio.NewReader(os.Stdin); var a, b int; fmt.Fscan(r, &a, &b); fmt.Println(a*b) }`
	sub2, res2, err := svc.Submit(ctx, userID, problemID, wrong, "go")
	if err != nil {
		t.Fatalf("submit wrong: %v", err)
	}
	if res2.Verdict != WA {
		t.Errorf("verdict = %s, want WA", res2.Verdict)
	}
	if sub2.Status != store.SubmissionStatusAttempted {
		t.Errorf("status = %s, want attempted", sub2.Status)
	}
	if rec.n != 1 {
		t.Errorf("WA must not record a solve activity; n = %d", rec.n)
	}
}
