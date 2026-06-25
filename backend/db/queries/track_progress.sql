-- name: CompletedVideoIDs :many
SELECT video_id FROM video_progress
WHERE user_id = sqlc.arg('user_id') AND completed = true
  AND video_id = ANY(sqlc.arg('ids')::uuid[]);

-- name: ReadArticleIDs :many
SELECT article_id FROM article_reads
WHERE user_id = sqlc.arg('user_id')
  AND article_id = ANY(sqlc.arg('ids')::uuid[]);

-- name: PassedQuizIDs :many
SELECT DISTINCT quiz_id FROM quiz_attempts
WHERE user_id = sqlc.arg('user_id') AND passed = true
  AND quiz_id = ANY(sqlc.arg('ids')::uuid[]);

-- name: SolvedProblemIDs :many
SELECT DISTINCT problem_id FROM problem_submissions
WHERE user_id = sqlc.arg('user_id') AND status = 'solved'
  AND problem_id = ANY(sqlc.arg('ids')::uuid[]);
