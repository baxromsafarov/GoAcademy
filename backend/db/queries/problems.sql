-- name: ListProblems :many
SELECT * FROM problems
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
ORDER BY created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountProblems :one
SELECT count(*) FROM problems
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags));

-- name: GetProblemBySlug :one
SELECT * FROM problems WHERE slug = $1;
