-- name: ListCheatsheets :many
SELECT * FROM cheatsheets
WHERE (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category'))
  AND (sqlc.narg('language')::locale IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('q')::text IS NULL
       OR title ILIKE '%' || sqlc.narg('q') || '%'
       OR category ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY category, title
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountCheatsheets :one
SELECT count(*) FROM cheatsheets
WHERE (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category'))
  AND (sqlc.narg('language')::locale IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('q')::text IS NULL
       OR title ILIKE '%' || sqlc.narg('q') || '%'
       OR category ILIKE '%' || sqlc.narg('q') || '%');

-- name: GetCheatsheetByID :one
SELECT * FROM cheatsheets WHERE id = $1;
