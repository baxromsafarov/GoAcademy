-- name: MarkArticleRead :one
-- Idempotent: returns the new row on first read, or no rows if already read.
INSERT INTO article_reads (user_id, article_id)
VALUES ($1, $2)
ON CONFLICT (user_id, article_id) DO NOTHING
RETURNING *;

-- name: GetArticleRead :one
SELECT * FROM article_reads WHERE user_id = $1 AND article_id = $2;
