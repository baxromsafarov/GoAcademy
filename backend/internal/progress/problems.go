package progress

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

const maxSubmissionCodeBytes = 64 * 1024 // 64 KiB

// SubmitProblemInput is a problem submission. In the MVP the verdict is set by
// the user ("mark as solved"); CHAPTER 17 replaces this with an automatic judge.
type SubmitProblemInput struct {
	Code     string
	Language string
	Solved   bool
}

// ProblemSubmissionResult carries the saved submission and, once the problem is
// solved, the revealed reference solution.
type ProblemSubmissionResult struct {
	Submission        store.ProblemSubmission
	ReferenceSolution string // populated only when the problem is solved
}

// SubmitProblem saves a submission for the problem identified by slug. Marking it
// solved reveals the reference solution and records a "problem_solved" activity
// the first time the user solves it.
func (s *Service) SubmitProblem(ctx context.Context, userID, slug string, in SubmitProblemInput) (ProblemSubmissionResult, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return ProblemSubmissionResult{}, apierr.Unauthorized("invalid user")
	}
	if len(in.Code) > maxSubmissionCodeBytes {
		return ProblemSubmissionResult{}, apierr.Validation("code is too large (max 64 KiB)")
	}
	language := strings.TrimSpace(in.Language)
	if language == "" {
		language = "go"
	}

	problem, err := s.resolveProblem(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProblemSubmissionResult{}, apierr.NotFound("problem not found")
	}
	if err != nil {
		return ProblemSubmissionResult{}, err
	}

	alreadySolved, err := s.queries.HasSolvedProblem(ctx, store.HasSolvedProblemParams{UserID: uid, ProblemID: problem.ID})
	if err != nil {
		return ProblemSubmissionResult{}, err
	}

	status := store.SubmissionStatusAttempted
	if in.Solved {
		status = store.SubmissionStatusSolved
	}

	sub, err := s.queries.CreateProblemSubmission(ctx, store.CreateProblemSubmissionParams{
		UserID:    uid,
		ProblemID: problem.ID,
		Status:    status,
		Code:      in.Code,
		Language:  language,
	})
	if err != nil {
		return ProblemSubmissionResult{}, err
	}

	if !alreadySolved && status == store.SubmissionStatusSolved {
		_ = s.activity.Record(ctx, activity.Event{
			UserID:  userID,
			Type:    "problem_solved",
			RefType: "problem",
			RefID:   pgxutil.UUIDString(problem.ID),
		})
	}

	result := ProblemSubmissionResult{Submission: sub}
	if alreadySolved || status == store.SubmissionStatusSolved {
		result.ReferenceSolution = problem.ReferenceSolutionMarkdown
	}
	return result, nil
}

// GetProblemSolution returns the reference solution, but only if the user has a
// solved submission for the problem; otherwise it returns 403.
func (s *Service) GetProblemSolution(ctx context.Context, userID, slug string) (string, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return "", apierr.Unauthorized("invalid user")
	}

	problem, err := s.resolveProblem(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", apierr.NotFound("problem not found")
	}
	if err != nil {
		return "", err
	}

	solved, err := s.queries.HasSolvedProblem(ctx, store.HasSolvedProblemParams{UserID: uid, ProblemID: problem.ID})
	if err != nil {
		return "", err
	}
	if !solved {
		return "", apierr.Forbidden("solve the problem to reveal the reference solution")
	}
	return problem.ReferenceSolutionMarkdown, nil
}
