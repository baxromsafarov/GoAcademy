-- name: ListTracks :many
SELECT * FROM tracks
WHERE (sqlc.narg('level')::difficulty IS NULL OR level = sqlc.narg('level'))
  AND (sqlc.narg('language')::locale  IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('q')::text           IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY position, created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountTracks :one
SELECT count(*) FROM tracks
WHERE (sqlc.narg('level')::difficulty IS NULL OR level = sqlc.narg('level'))
  AND (sqlc.narg('language')::locale  IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('q')::text           IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%');

-- name: GetTrackByID :one
SELECT * FROM tracks WHERE id = $1;

-- name: ListTrackItems :many
SELECT * FROM track_items WHERE track_id = $1 ORDER BY position, id;
