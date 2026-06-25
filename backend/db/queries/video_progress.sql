-- name: GetVideoProgress :one
SELECT * FROM video_progress WHERE user_id = $1 AND video_id = $2;

-- name: UpsertVideoProgress :one
-- watched_percent never decreases; completed is sticky (once true, stays true).
INSERT INTO video_progress (user_id, video_id, watched_percent, last_position_seconds, completed)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, video_id) DO UPDATE SET
    watched_percent       = GREATEST(video_progress.watched_percent, EXCLUDED.watched_percent),
    last_position_seconds = EXCLUDED.last_position_seconds,
    completed             = video_progress.completed OR EXCLUDED.completed
RETURNING *;
