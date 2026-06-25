-- name: CreateNote :one
INSERT INTO notes (user_id, content_type, content_id, body)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateNote :one
-- Owner-scoped: a non-owner (or missing id) matches no row.
UPDATE notes SET body = sqlc.arg('body')
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING *;

-- name: DeleteNote :execrows
DELETE FROM notes WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id');

-- name: ListUserNotes :many
SELECT * FROM notes WHERE user_id = $1 ORDER BY created_at DESC, id;
