-- name: ListQuizzes :many
SELECT * FROM quizzes
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
  AND (sqlc.narg('q')::text              IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.arg('show_hidden')::bool OR NOT ('hidden' = ANY(tags)))
ORDER BY created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountQuizzes :one
SELECT count(*) FROM quizzes
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
  AND (sqlc.narg('q')::text              IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.arg('show_hidden')::bool OR NOT ('hidden' = ANY(tags)));

-- name: GetQuizByID :one
SELECT * FROM quizzes WHERE id = $1;

-- name: ListQuizQuestions :many
SELECT * FROM quiz_questions WHERE quiz_id = $1 ORDER BY position, id;

-- name: ListQuizOptionsByQuiz :many
SELECT o.* FROM quiz_options o
JOIN quiz_questions q ON q.id = o.question_id
WHERE q.quiz_id = $1
ORDER BY q.position, o.position, o.id;
