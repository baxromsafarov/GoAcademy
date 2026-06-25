package quiz

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

type countRecorder struct{ n int }

func (c *countRecorder) Record(context.Context, activity.Event) error { c.n++; return nil }

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run quiz integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func scanID(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&id); err != nil {
		t.Fatalf("seed (%s): %v", sql, err)
	}
	return id
}

func reviewFor(res AttemptResult, qid string) (QuestionReview, bool) {
	for _, r := range res.Review {
		if r.QuestionID == qid {
			return r, true
		}
	}
	return QuestionReview{}, false
}

func TestService_SubmitAttempt_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := &countRecorder{}
	svc := NewService(pool, content.NewService(pool), rec)

	email := fmt.Sprintf("quiz-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(ctx, store.CreateUserParams{
		Email: email, PasswordHash: "x", DisplayName: "Q", Locale: store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	userID := pgxutil.UUIDString(u.ID)

	marker := fmt.Sprintf("qzsub-%d", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM quizzes WHERE $1 = ANY(tags)", marker) })

	quizID := scanID(t, pool, "INSERT INTO quizzes (title, pass_threshold, tags) VALUES ('Quiz', 50, ARRAY[$1]) RETURNING id::text", marker)
	q1 := scanID(t, pool, "INSERT INTO quiz_questions (quiz_id, prompt, type, position) VALUES ($1,'q1','single',1) RETURNING id::text", quizID)
	q2 := scanID(t, pool, "INSERT INTO quiz_questions (quiz_id, prompt, type, position) VALUES ($1,'q2','multiple',2) RETURNING id::text", quizID)
	a := scanID(t, pool, "INSERT INTO quiz_options (question_id,text,is_correct,position) VALUES ($1,'a',true,1)  RETURNING id::text", q1)
	b := scanID(t, pool, "INSERT INTO quiz_options (question_id,text,is_correct,position) VALUES ($1,'b',false,2) RETURNING id::text", q1)
	x := scanID(t, pool, "INSERT INTO quiz_options (question_id,text,is_correct,position) VALUES ($1,'x',true,1)  RETURNING id::text", q2)
	y := scanID(t, pool, "INSERT INTO quiz_options (question_id,text,is_correct,position) VALUES ($1,'y',true,2)  RETURNING id::text", q2)
	z := scanID(t, pool, "INSERT INTO quiz_options (question_id,text,is_correct,position) VALUES ($1,'z',false,3) RETURNING id::text", q2)

	// All correct -> 100, passed; activity fired once.
	res, err := svc.Submit(ctx, userID, quizID, SubmitInput{Answers: map[string][]string{q1: {a}, q2: {x, y}}})
	if err != nil {
		t.Fatalf("Submit (all correct): %v", err)
	}
	if res.Score != 100 || !res.Passed {
		t.Errorf("all correct: score=%d passed=%v, want 100/true", res.Score, res.Passed)
	}
	if res.AttemptID == "" {
		t.Error("attempt should be persisted (empty id)")
	}
	if rec.n != 1 {
		t.Errorf("activity n=%d, want 1", rec.n)
	}
	// Review reveals correct answers.
	if r, ok := reviewFor(res, q2); !ok || !r.Correct || len(r.CorrectIDs) != 2 {
		t.Errorf("q2 review wrong: %+v", r)
	}

	// Half correct (q2 partial) -> 50, passed (>= threshold 50).
	half, _ := svc.Submit(ctx, userID, quizID, SubmitInput{Answers: map[string][]string{q1: {a}, q2: {x}}})
	if half.Score != 50 || !half.Passed {
		t.Errorf("half: score=%d passed=%v, want 50/true", half.Score, half.Passed)
	}

	// All wrong -> 0, not passed.
	wrong, _ := svc.Submit(ctx, userID, quizID, SubmitInput{Answers: map[string][]string{q1: {b}, q2: {z}}})
	if wrong.Score != 0 || wrong.Passed {
		t.Errorf("wrong: score=%d passed=%v, want 0/false", wrong.Score, wrong.Passed)
	}

	// Validation: single-choice with two options.
	if _, err := svc.Submit(ctx, userID, quizID, SubmitInput{Answers: map[string][]string{q1: {a, b}}}); err == nil {
		t.Error("single-choice with two options should be rejected")
	}
	// Validation: unknown option id.
	if _, err := svc.Submit(ctx, userID, quizID, SubmitInput{Answers: map[string][]string{q1: {"00000000-0000-0000-0000-000000000000"}}}); err == nil {
		t.Error("unknown option should be rejected")
	}
	// Unknown quiz -> not found.
	if _, err := svc.Submit(ctx, userID, "00000000-0000-0000-0000-000000000000", SubmitInput{Answers: map[string][]string{}}); err == nil {
		t.Error("unknown quiz should be not found")
	}
}
