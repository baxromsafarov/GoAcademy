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
-- The saved content's title is resolved polymorphically so the UI can show what
-- was bookmarked (content_id is a uuid across every content type).
SELECT b.*,
  COALESCE(v.title, a.title, q.title, p.title, mp.title, t.title, cs.title, '') AS title
FROM bookmarks b
LEFT JOIN videos v         ON b.content_type = 'video'      AND v.id  = b.content_id
LEFT JOIN articles a       ON b.content_type = 'article'    AND a.id  = b.content_id
LEFT JOIN quizzes q        ON b.content_type = 'quiz'       AND q.id  = b.content_id
LEFT JOIN problems p       ON b.content_type = 'problem'    AND p.id  = b.content_id
LEFT JOIN mini_projects mp ON b.content_type = 'project'    AND mp.id = b.content_id
LEFT JOIN tracks t         ON b.content_type = 'track'      AND t.id  = b.content_id
LEFT JOIN cheatsheets cs   ON b.content_type = 'cheatsheet' AND cs.id = b.content_id
WHERE b.user_id = $1
ORDER BY b.created_at DESC, b.id;
