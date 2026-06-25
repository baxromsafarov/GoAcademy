-- name: InsertActivity :one
INSERT INTO activity_log (user_id, activity_type, ref_type, ref_id, xp_earned)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListUserActivity :many
SELECT * FROM activity_log
WHERE user_id = $1
ORDER BY occurred_at DESC, id;

-- name: CountUserActivityByType :one
SELECT count(*) FROM activity_log
WHERE user_id = $1 AND activity_type = $2;
