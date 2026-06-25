-- name: ListUsers :many
-- Optional case-insensitive search over email/display_name, paginated.
SELECT * FROM users
WHERE (sqlc.narg('q')::text IS NULL
       OR email ILIKE '%' || sqlc.narg('q') || '%'
       OR display_name ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY created_at DESC, id
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountUsers :one
SELECT count(*) FROM users
WHERE (sqlc.narg('q')::text IS NULL
       OR email ILIKE '%' || sqlc.narg('q') || '%'
       OR display_name ILIKE '%' || sqlc.narg('q') || '%');

-- name: AdminUpdateUser :one
UPDATE users SET role = $2, is_blocked = $3 WHERE id = $1 RETURNING *;
