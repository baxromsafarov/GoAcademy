package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// QuizOptionInput is one answer option.
type QuizOptionInput struct {
	Text      string
	IsCorrect bool
}

// QuizQuestionInput is one question with its options.
type QuizQuestionInput struct {
	Prompt  string
	Type    string // single | multiple
	Options []QuizOptionInput
}

// QuizInput is the create/update payload for a quiz and its full question set.
type QuizInput struct {
	Title         string
	Description   string
	PassThreshold int
	Difficulty    string
	Language      string
	Tags          []string
	Questions     []QuizQuestionInput
}

func (in QuizInput) validate() error {
	details := map[string]string{}
	validateMeta(details, in.Title, in.Difficulty, in.Language)
	if in.PassThreshold < 0 || in.PassThreshold > 100 {
		details["pass_threshold"] = "must be between 0 and 100"
	}
	if len(in.Questions) == 0 {
		details["questions"] = "must have at least one question"
	}
	for i, q := range in.Questions {
		key := fmt.Sprintf("questions[%d]", i)
		if strings.TrimSpace(q.Prompt) == "" {
			details[key+".prompt"] = "must not be empty"
		}
		if q.Type != "single" && q.Type != "multiple" {
			details[key+".type"] = "must be single or multiple"
		}
		if len(q.Options) < 2 {
			details[key+".options"] = "must have at least two options"
		}
		correct := 0
		for _, o := range q.Options {
			if o.IsCorrect {
				correct++
			}
		}
		switch {
		case correct == 0:
			details[key+".options"] = "must have at least one correct option"
		case q.Type == "single" && correct > 1:
			details[key+".options"] = "single-choice must have exactly one correct option"
		}
	}
	if len(details) > 0 {
		return apierr.Validation("invalid quiz").WithDetails(details)
	}
	return nil
}

// CreateQuiz inserts a quiz with its questions and options atomically.
func (s *Service) CreateQuiz(ctx context.Context, in QuizInput) (store.Quiz, error) {
	if err := in.validate(); err != nil {
		return store.Quiz{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.Quiz{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	quiz, err := q.CreateQuiz(ctx, store.CreateQuizParams{
		Title: in.Title, Description: in.Description, PassThreshold: int32(in.PassThreshold),
		Difficulty: store.Difficulty(in.Difficulty), Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	if err != nil {
		return store.Quiz{}, err
	}
	if err := insertQuizQuestions(ctx, q, quiz.ID, in.Questions); err != nil {
		return store.Quiz{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.Quiz{}, err
	}
	return quiz, nil
}

// UpdateQuiz replaces a quiz's metadata and its full question set atomically.
func (s *Service) UpdateQuiz(ctx context.Context, id string, in QuizInput) (store.Quiz, error) {
	qid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Quiz{}, apierr.NotFound("quiz not found")
	}
	if err := in.validate(); err != nil {
		return store.Quiz{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.Quiz{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	quiz, err := q.UpdateQuiz(ctx, store.UpdateQuizParams{
		ID: qid, Title: in.Title, Description: in.Description, PassThreshold: int32(in.PassThreshold),
		Difficulty: store.Difficulty(in.Difficulty), Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Quiz{}, apierr.NotFound("quiz not found")
	}
	if err != nil {
		return store.Quiz{}, err
	}
	if err := q.DeleteQuizQuestions(ctx, qid); err != nil {
		return store.Quiz{}, err
	}
	if err := insertQuizQuestions(ctx, q, qid, in.Questions); err != nil {
		return store.Quiz{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.Quiz{}, err
	}
	return quiz, nil
}

func insertQuizQuestions(ctx context.Context, q *store.Queries, quizID pgtype.UUID, questions []QuizQuestionInput) error {
	for i, qq := range questions {
		question, err := q.CreateQuizQuestion(ctx, store.CreateQuizQuestionParams{
			QuizID: quizID, Prompt: qq.Prompt, Type: store.QuizQuestionType(qq.Type), Position: int32(i + 1),
		})
		if err != nil {
			return err
		}
		for j, opt := range qq.Options {
			if _, err := q.CreateQuizOption(ctx, store.CreateQuizOptionParams{
				QuestionID: question.ID, Text: opt.Text, IsCorrect: opt.IsCorrect, Position: int32(j + 1),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteQuiz removes a quiz (cascading its questions and options).
func (s *Service) DeleteQuiz(ctx context.Context, id string) error {
	qid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("quiz not found")
	}
	n, err := s.queries.DeleteQuiz(ctx, qid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("quiz not found")
	}
	return nil
}
