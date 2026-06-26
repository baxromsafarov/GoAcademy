-- name: ListVideos :many
SELECT * FROM videos
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
  AND (sqlc.narg('q')::text              IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.arg('show_hidden')::bool OR NOT ('hidden' = ANY(tags)))
ORDER BY created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountVideos :one
SELECT count(*) FROM videos
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
  AND (sqlc.narg('q')::text              IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.arg('show_hidden')::bool OR NOT ('hidden' = ANY(tags)));

-- name: GetVideoByID :one
SELECT * FROM videos WHERE id = $1;
