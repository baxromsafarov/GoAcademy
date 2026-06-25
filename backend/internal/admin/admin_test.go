package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run admin integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func marker() string { return fmt.Sprintf("adm-%d", time.Now().UnixNano()) }

func TestAdminVideoCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	q := store.New(pool)

	v, err := svc.CreateVideo(ctx, VideoInput{
		Title: "Intro", Description: "d", YoutubeID: "yt1", DurationSeconds: 60,
		Difficulty: "beginner", Language: "en", Tags: []string{"go"},
	})
	if err != nil {
		t.Fatalf("CreateVideo: %v", err)
	}
	id := pgxutil.UUIDString(v.ID)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM videos WHERE id = $1", v.ID) })

	v2, err := svc.UpdateVideo(ctx, id, VideoInput{
		Title: "Intro v2", YoutubeID: "yt1", DurationSeconds: 90, Difficulty: "advanced", Language: "ru",
	})
	if err != nil || v2.Title != "Intro v2" || string(v2.Difficulty) != "advanced" {
		t.Fatalf("UpdateVideo = (%+v, %v)", v2, err)
	}
	if got, _ := q.GetVideoByID(ctx, v.ID); got.Title != "Intro v2" {
		t.Errorf("read back title = %q", got.Title)
	}

	if err := svc.DeleteVideo(ctx, id); err != nil {
		t.Fatalf("DeleteVideo: %v", err)
	}
	if _, err := q.GetVideoByID(ctx, v.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("video should be gone, got %v", err)
	}
	if err := svc.DeleteVideo(ctx, id); err == nil {
		t.Error("deleting a missing video should be not-found")
	}

	// Validation.
	if _, err := svc.CreateVideo(ctx, VideoInput{Title: "", YoutubeID: "y", Difficulty: "beginner", Language: "en"}); err == nil {
		t.Error("empty title should fail")
	}
	if _, err := svc.CreateVideo(ctx, VideoInput{Title: "x", YoutubeID: "y", Difficulty: "bogus", Language: "en"}); err == nil {
		t.Error("bad difficulty should fail")
	}
}

func TestAdminArticleCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	m := marker()
	slug := "art-" + m

	a, err := svc.CreateArticle(ctx, ArticleInput{Title: "A", Slug: slug, BodyMarkdown: "b", Difficulty: "beginner", Language: "en"})
	if err != nil {
		t.Fatalf("CreateArticle: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM articles WHERE id = $1", a.ID) })

	// Duplicate slug -> conflict.
	if _, err := svc.CreateArticle(ctx, ArticleInput{Title: "B", Slug: slug, Difficulty: "beginner", Language: "en"}); err == nil {
		t.Error("duplicate slug should conflict")
	}

	a2, err := svc.UpdateArticle(ctx, pgxutil.UUIDString(a.ID), ArticleInput{Title: "A2", Slug: slug, BodyMarkdown: "b2", Difficulty: "intermediate", Language: "ru"})
	if err != nil || a2.Title != "A2" {
		t.Fatalf("UpdateArticle = (%+v, %v)", a2, err)
	}
	if err := svc.DeleteArticle(ctx, pgxutil.UUIDString(a.ID)); err != nil {
		t.Fatalf("DeleteArticle: %v", err)
	}
}

func TestAdminQuizCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	q := store.New(pool)

	in := QuizInput{
		Title: "Quiz", Difficulty: "beginner", Language: "en", PassThreshold: 50,
		Questions: []QuizQuestionInput{
			{Prompt: "q1", Type: "single", Options: []QuizOptionInput{{Text: "a", IsCorrect: true}, {Text: "b"}}},
			{Prompt: "q2", Type: "multiple", Options: []QuizOptionInput{{Text: "c", IsCorrect: true}, {Text: "d", IsCorrect: true}}},
		},
	}
	quiz, err := svc.CreateQuiz(ctx, in)
	if err != nil {
		t.Fatalf("CreateQuiz: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM quizzes WHERE id = $1", quiz.ID) })

	if qs, _ := q.ListQuizQuestions(ctx, quiz.ID); len(qs) != 2 {
		t.Fatalf("want 2 questions, got %d", len(qs))
	}

	// Update replaces the question set with a single question.
	in.Title = "Quiz v2"
	in.Questions = in.Questions[:1]
	q2, err := svc.UpdateQuiz(ctx, pgxutil.UUIDString(quiz.ID), in)
	if err != nil || q2.Title != "Quiz v2" {
		t.Fatalf("UpdateQuiz = (%+v, %v)", q2, err)
	}
	if qs, _ := q.ListQuizQuestions(ctx, quiz.ID); len(qs) != 1 {
		t.Errorf("update should replace questions, got %d", len(qs))
	}

	// Validation: a quiz needs questions; single-choice needs exactly one correct.
	if _, err := svc.CreateQuiz(ctx, QuizInput{Title: "x", Difficulty: "beginner", Language: "en"}); err == nil {
		t.Error("quiz with no questions should fail")
	}
	bad := QuizInput{Title: "x", Difficulty: "beginner", Language: "en", Questions: []QuizQuestionInput{
		{Prompt: "p", Type: "single", Options: []QuizOptionInput{{Text: "a", IsCorrect: true}, {Text: "b", IsCorrect: true}}},
	}}
	if _, err := svc.CreateQuiz(ctx, bad); err == nil {
		t.Error("single-choice with two correct options should fail")
	}

	if err := svc.DeleteQuiz(ctx, pgxutil.UUIDString(quiz.ID)); err != nil {
		t.Fatalf("DeleteQuiz: %v", err)
	}
	if qs, _ := q.ListQuizQuestions(ctx, quiz.ID); len(qs) != 0 {
		t.Errorf("questions should cascade-delete, got %d", len(qs))
	}
}

func TestAdminProblemCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	m := marker()
	slug := "prob-" + m

	caseCount := func(problemID pgtype.UUID) int {
		var n int
		_ = pool.QueryRow(ctx, "SELECT count(*) FROM problem_test_cases WHERE problem_id = $1", problemID).Scan(&n)
		return n
	}

	prob, err := svc.CreateProblem(ctx, ProblemInput{
		Title: "P", Slug: slug, StatementMarkdown: "do x", Difficulty: "beginner", Language: "en",
		ReferenceSolutionMarkdown: "answer", SampleIO: json.RawMessage(`[{"in":"1","out":"2"}]`),
		TestCases: []TestCaseInput{{Input: "1", ExpectedOutput: "2", IsSample: true}, {Input: "3", ExpectedOutput: "4"}},
	})
	if err != nil {
		t.Fatalf("CreateProblem: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM problems WHERE id = $1", prob.ID) })
	if n := caseCount(prob.ID); n != 2 {
		t.Fatalf("want 2 test cases, got %d", n)
	}

	// Update replaces test cases (now one).
	p2, err := svc.UpdateProblem(ctx, pgxutil.UUIDString(prob.ID), ProblemInput{
		Title: "P2", Slug: slug, StatementMarkdown: "do y", Difficulty: "advanced", Language: "ru",
		TestCases: []TestCaseInput{{Input: "5", ExpectedOutput: "6"}},
	})
	if err != nil || p2.Title != "P2" {
		t.Fatalf("UpdateProblem = (%+v, %v)", p2, err)
	}
	if n := caseCount(prob.ID); n != 1 {
		t.Errorf("update should replace test cases, got %d", n)
	}

	// Invalid sample_io.
	if _, err := svc.CreateProblem(ctx, ProblemInput{Title: "Q", Slug: "q-" + m, Difficulty: "beginner", Language: "en", SampleIO: json.RawMessage(`{bad`)}); err == nil {
		t.Error("invalid sample_io should fail validation")
	}

	if err := svc.DeleteProblem(ctx, pgxutil.UUIDString(prob.ID)); err != nil {
		t.Fatalf("DeleteProblem: %v", err)
	}
}
