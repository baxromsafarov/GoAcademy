package judge

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/runner"
	"github.com/goacademy/backend/internal/store"
)

// ErrNoTestCases means a problem has no test cases to judge against; callers may
// fall back to manual marking.
var ErrNoTestCases = errors.New("problem has no test cases")

// Service judges a submission against a problem's stored test cases and records
// the result (with its verdict) as a problem_submissions row.
type Service struct {
	judge    *Judge
	queries  *store.Queries
	activity activity.Recorder
}

// NewService wires the judge to the database and the activity recorder.
func NewService(r *runner.Runner, pool *pgxpool.Pool, rec activity.Recorder, limits runner.Limits) *Service {
	return &Service{judge: New(r, limits), queries: store.New(pool), activity: rec}
}

// Submit grades code for problemID, persists the judged submission, and returns
// it together with the result. On an OK verdict it records a problem_solved
// activity (idempotent in the recorder) and marks the submission solved.
func (s *Service) Submit(ctx context.Context, userID, problemID, code, language string) (store.ProblemSubmission, Result, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return store.ProblemSubmission{}, Result{}, apierr.Unauthorized("invalid user")
	}
	pid, err := pgxutil.ParseUUID(problemID)
	if err != nil {
		return store.ProblemSubmission{}, Result{}, apierr.NotFound("problem not found")
	}

	rows, err := s.queries.ListProblemTestCases(ctx, pid)
	if err != nil {
		return store.ProblemSubmission{}, Result{}, err
	}
	if len(rows) == 0 {
		return store.ProblemSubmission{}, Result{}, ErrNoTestCases
	}

	cases := make([]TestCase, 0, len(rows))
	for _, tc := range rows {
		cases = append(cases, TestCase{Input: tc.Input, ExpectedOutput: tc.ExpectedOutput, IsSample: tc.IsSample})
	}

	result, err := s.judge.Run(ctx, code, cases)
	if err != nil {
		return store.ProblemSubmission{}, Result{}, err
	}

	status := store.SubmissionStatusAttempted
	if result.Verdict == OK {
		status = store.SubmissionStatusSolved
	}
	verdictJSON, err := json.Marshal(result)
	if err != nil {
		return store.ProblemSubmission{}, Result{}, err
	}

	sub, err := s.queries.CreateJudgedSubmission(ctx, store.CreateJudgedSubmissionParams{
		UserID:    uid,
		ProblemID: pid,
		Status:    status,
		Code:      code,
		Language:  language,
		Verdict:   verdictJSON,
	})
	if err != nil {
		return store.ProblemSubmission{}, Result{}, err
	}

	if result.Verdict == OK {
		// Best-effort: a recorder failure must not fail the submission write.
		_ = s.activity.Record(ctx, activity.Event{
			UserID:  userID,
			Type:    "problem_solved",
			RefType: "problem",
			RefID:   problemID,
		})
	}
	return sub, result, nil
}
