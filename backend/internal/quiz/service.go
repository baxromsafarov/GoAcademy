package quiz

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// Service grades quiz submissions and records attempts.
type Service struct {
	queries  *store.Queries
	content  *content.Service
	activity activity.Recorder
}

// NewService wires the quiz service. It reuses content.Service to load quizzes.
func NewService(pool *pgxpool.Pool, content *content.Service, rec activity.Recorder) *Service {
	return &Service{queries: store.New(pool), content: content, activity: rec}
}

// SubmitInput is a quiz submission: a map of question id to selected option ids.
type SubmitInput struct {
	Answers map[string][]string
}

// QuestionReview reveals, per question, whether the answer was correct and which
// options were correct (shown only after submission).
type QuestionReview struct {
	QuestionID string
	Correct    bool
	CorrectIDs []string
}

// AttemptResult is the graded outcome of a submission.
type AttemptResult struct {
	AttemptID string
	Score     int
	Passed    bool
	Review    []QuestionReview
	CreatedAt time.Time
}

// Submit grades the submission against quizID, persists the attempt, records
// activity and returns the score, pass/fail and a per-question review.
func (s *Service) Submit(ctx context.Context, userID, quizID string, in SubmitInput) (AttemptResult, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return AttemptResult{}, apierr.Unauthorized("invalid user")
	}

	detail, err := s.content.GetQuizDetail(ctx, quizID) // 404 if the quiz does not exist
	if err != nil {
		return AttemptResult{}, err
	}

	questions := make([]Question, 0, len(detail.Questions))
	byID := make(map[string]content.QuizQuestionDetail, len(detail.Questions))
	optionOf := make(map[string]map[string]struct{}, len(detail.Questions))
	for _, qd := range detail.Questions {
		qid := pgxutil.UUIDString(qd.Question.ID)
		correct := make([]string, 0)
		opts := make(map[string]struct{}, len(qd.Options))
		for _, o := range qd.Options {
			oid := pgxutil.UUIDString(o.ID)
			opts[oid] = struct{}{}
			if o.IsCorrect {
				correct = append(correct, oid)
			}
		}
		questions = append(questions, Question{ID: qid, CorrectIDs: correct})
		byID[qid] = qd
		optionOf[qid] = opts
	}

	if err := validateSubmission(in.Answers, byID, optionOf); err != nil {
		return AttemptResult{}, err
	}

	scored := Score(questions, in.Answers)
	passed := scored.Score >= int(detail.Quiz.PassThreshold)

	answersJSON, err := json.Marshal(normalizeAnswers(in.Answers))
	if err != nil {
		return AttemptResult{}, err
	}

	attempt, err := s.queries.CreateQuizAttempt(ctx, store.CreateQuizAttemptParams{
		UserID:  uid,
		QuizID:  detail.Quiz.ID,
		Score:   int32(scored.Score),
		Passed:  passed,
		Answers: answersJSON,
	})
	if err != nil {
		return AttemptResult{}, err
	}

	evType := "quiz_attempt"
	if passed {
		evType = "quiz_passed"
	}
	_ = s.activity.Record(ctx, activity.Event{UserID: userID, Type: evType, RefType: "quiz", RefID: quizID})

	review := make([]QuestionReview, 0, len(scored.Results))
	for _, r := range scored.Results {
		review = append(review, QuestionReview{QuestionID: r.QuestionID, Correct: r.Correct, CorrectIDs: r.CorrectIDs})
	}
	return AttemptResult{
		AttemptID: pgxutil.UUIDString(attempt.ID),
		Score:     scored.Score,
		Passed:    passed,
		Review:    review,
		CreatedAt: attempt.CreatedAt.Time,
	}, nil
}

// validateSubmission rejects unknown questions/options and multi-selections on
// single-choice questions.
func validateSubmission(answers map[string][]string, byID map[string]content.QuizQuestionDetail, optionOf map[string]map[string]struct{}) error {
	details := map[string]string{}
	for qid, selected := range answers {
		qd, ok := byID[qid]
		if !ok {
			details[qid] = "unknown question for this quiz"
			continue
		}
		if qd.Question.Type == store.QuizQuestionTypeSingle && len(dedup(selected)) > 1 {
			details[qid] = "single-choice question accepts at most one option"
			continue
		}
		for _, oid := range selected {
			if _, ok := optionOf[qid][oid]; !ok {
				details[qid] = "contains an option that does not belong to this question"
				break
			}
		}
	}
	if len(details) > 0 {
		return apierr.Validation("invalid answers").WithDetails(details)
	}
	return nil
}

func dedup(xs []string) []string {
	seen := make(map[string]struct{}, len(xs))
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			out = append(out, x)
		}
	}
	return out
}

func normalizeAnswers(answers map[string][]string) map[string][]string {
	out := make(map[string][]string, len(answers))
	for k, v := range answers {
		out[k] = dedup(v)
	}
	return out
}
