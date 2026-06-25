package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// TestCaseInput is one judge test case (used by the online judge in CH17).
type TestCaseInput struct {
	Input          string
	ExpectedOutput string
	IsSample       bool
}

// ProblemInput is the create/update payload for a problem and its test cases.
type ProblemInput struct {
	Title                     string
	Slug                      string
	StatementMarkdown         string
	Difficulty                string
	Language                  string
	ReferenceSolutionMarkdown string
	SampleIO                  json.RawMessage
	Tags                      []string
	TestCases                 []TestCaseInput
}

func (in ProblemInput) validate() error {
	details := map[string]string{}
	validateMeta(details, in.Title, in.Difficulty, in.Language)
	if strings.TrimSpace(in.Slug) == "" {
		details["slug"] = "must not be empty"
	}
	if len(in.SampleIO) > 0 && !json.Valid(in.SampleIO) {
		details["sample_io"] = "must be valid JSON"
	}
	if len(details) > 0 {
		return apierr.Validation("invalid problem").WithDetails(details)
	}
	return nil
}

// sampleIOBytes defaults an empty sample_io to an empty JSON array.
func sampleIOBytes(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("[]")
	}
	return raw
}

// CreateProblem inserts a problem and its test cases atomically.
func (s *Service) CreateProblem(ctx context.Context, in ProblemInput) (store.Problem, error) {
	if err := in.validate(); err != nil {
		return store.Problem{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.Problem{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	prob, err := q.CreateProblem(ctx, store.CreateProblemParams{
		Title: in.Title, Slug: in.Slug, StatementMarkdown: in.StatementMarkdown,
		Difficulty: store.Difficulty(in.Difficulty), ReferenceSolutionMarkdown: in.ReferenceSolutionMarkdown,
		SampleIo: sampleIOBytes(in.SampleIO), Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	if isUniqueViolation(err) {
		return store.Problem{}, apierr.Conflict("a problem with this slug already exists")
	}
	if err != nil {
		return store.Problem{}, err
	}
	if err := insertTestCases(ctx, q, prob.ID, in.TestCases); err != nil {
		return store.Problem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.Problem{}, err
	}
	return prob, nil
}

// UpdateProblem replaces a problem's fields and its test cases atomically.
func (s *Service) UpdateProblem(ctx context.Context, id string, in ProblemInput) (store.Problem, error) {
	pid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Problem{}, apierr.NotFound("problem not found")
	}
	if err := in.validate(); err != nil {
		return store.Problem{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.Problem{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	prob, err := q.UpdateProblem(ctx, store.UpdateProblemParams{
		ID: pid, Title: in.Title, Slug: in.Slug, StatementMarkdown: in.StatementMarkdown,
		Difficulty: store.Difficulty(in.Difficulty), ReferenceSolutionMarkdown: in.ReferenceSolutionMarkdown,
		SampleIo: sampleIOBytes(in.SampleIO), Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return store.Problem{}, apierr.NotFound("problem not found")
	case isUniqueViolation(err):
		return store.Problem{}, apierr.Conflict("a problem with this slug already exists")
	case err != nil:
		return store.Problem{}, err
	}
	if err := q.DeleteProblemTestCases(ctx, pid); err != nil {
		return store.Problem{}, err
	}
	if err := insertTestCases(ctx, q, pid, in.TestCases); err != nil {
		return store.Problem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.Problem{}, err
	}
	return prob, nil
}

func insertTestCases(ctx context.Context, q *store.Queries, problemID pgtype.UUID, cases []TestCaseInput) error {
	for i, tc := range cases {
		if _, err := q.CreateProblemTestCase(ctx, store.CreateProblemTestCaseParams{
			ProblemID: problemID, Input: tc.Input, ExpectedOutput: tc.ExpectedOutput,
			IsSample: tc.IsSample, Position: int32(i + 1),
		}); err != nil {
			return err
		}
	}
	return nil
}

// DeleteProblem removes a problem (cascading test cases and submissions).
func (s *Service) DeleteProblem(ctx context.Context, id string) error {
	pid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("problem not found")
	}
	n, err := s.queries.DeleteProblem(ctx, pid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("problem not found")
	}
	return nil
}
