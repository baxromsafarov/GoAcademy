-- name: ListArticles :many
SELECT * FROM articles
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
ORDER BY created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountArticles :one
SELECT count(*) FROM articles
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags));

-- name: GetArticleBySlug :one
SELECT * FROM articles WHERE slug = $1;
