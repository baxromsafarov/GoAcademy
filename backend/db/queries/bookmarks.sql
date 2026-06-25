-- name: CreateBookmark :one
-- Idempotent add: a duplicate (user, content) returns the existing row (the
-- DO UPDATE is a no-op that lets RETURNING fire on conflict).
INSERT INTO bookmarks (user_id, content_type, content_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, content_type, content_id)
DO UPDATE SET content_id = EXCLUDED.content_id
RETURNING *;

-- name: DeleteBookmark :execrows
DELETE FROM bookmarks WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id');

-- name: ListUserBookmarks :many
SELECT * FROM bookmarks WHERE user_id = $1 ORDER BY created_at DESC, id;
