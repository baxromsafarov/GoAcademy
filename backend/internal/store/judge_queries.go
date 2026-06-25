package store

// Hand-written store methods for the online judge (CHAPTER 17). They live outside
// the sqlc-generated files (which sqlc owns) but follow the same style and read
// q.db like any other Queries method: listing a problem's full test set and
// recording a judged submission together with its verdict JSON.

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const listProblemTestCases = `-- name: ListProblemTestCases :many
SELECT id, problem_id, input, expected_output, is_sample, position
FROM problem_test_cases
WHERE problem_id = $1
ORDER BY position
`

// ListProblemTestCases returns a problem's full test set ordered by position.
func (q *Queries) ListProblemTestCases(ctx context.Context, problemID pgtype.UUID) ([]ProblemTestCase, error) {
	rows, err := q.db.Query(ctx, listProblemTestCases, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ProblemTestCase{}
	for rows.Next() {
		var i ProblemTestCase
		if err := rows.Scan(&i.ID, &i.ProblemID, &i.Input, &i.ExpectedOutput, &i.IsSample, &i.Position); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const createJudgedSubmission = `-- name: CreateJudgedSubmission :one
INSERT INTO problem_submissions (user_id, problem_id, status, code, language, verdict)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, problem_id, status, code, language, verdict, created_at
`

// CreateJudgedSubmissionParams carries a submission plus its judge verdict JSON.
type CreateJudgedSubmissionParams struct {
	UserID    pgtype.UUID
	ProblemID pgtype.UUID
	Status    SubmissionStatus
	Code      string
	Language  string
	Verdict   []byte
}

// CreateJudgedSubmission records a submission together with its verdict JSON.
func (q *Queries) CreateJudgedSubmission(ctx context.Context, arg CreateJudgedSubmissionParams) (ProblemSubmission, error) {
	row := q.db.QueryRow(ctx, createJudgedSubmission,
		arg.UserID, arg.ProblemID, arg.Status, arg.Code, arg.Language, arg.Verdict)
	var i ProblemSubmission
	err := row.Scan(&i.ID, &i.UserID, &i.ProblemID, &i.Status, &i.Code, &i.Language, &i.Verdict, &i.CreatedAt)
	return i, err
}
