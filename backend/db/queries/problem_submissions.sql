-- name: CreateProblemSubmission :one
INSERT INTO problem_submissions (user_id, problem_id, status, code, language)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: HasSolvedProblem :one
SELECT EXISTS (
    SELECT 1 FROM problem_submissions
    WHERE user_id = $1 AND problem_id = $2 AND status = 'solved'
) AS solved;
