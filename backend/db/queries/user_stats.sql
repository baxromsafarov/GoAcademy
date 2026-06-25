-- name: GetUserStats :one
SELECT * FROM user_stats WHERE user_id = $1;

-- name: ActivityExists :one
-- True if the user already has an activity of this type for this referenced
-- entity (used to award XP only on the first occurrence).
SELECT EXISTS(
    SELECT 1 FROM activity_log
    WHERE user_id = sqlc.arg('user_id')
      AND activity_type = sqlc.arg('activity_type')
      AND ref_id = sqlc.arg('ref_id')
);

-- name: EnsureUserStats :exec
-- Guarantees a stats row exists (defaults) so it can then be locked and updated.
INSERT INTO user_stats (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING;

-- name: LockUserStats :one
-- Row-locks the user's stats for the rest of the transaction, serializing
-- concurrent XP/streak updates (no lost updates).
SELECT * FROM user_stats WHERE user_id = $1 FOR UPDATE;

-- name: UpdateUserStats :exec
UPDATE user_stats SET
    total_xp         = sqlc.arg('total_xp'),
    level            = sqlc.arg('level'),
    current_streak   = sqlc.arg('current_streak'),
    longest_streak   = sqlc.arg('longest_streak'),
    last_active_date = sqlc.arg('last_active_date')
WHERE user_id = sqlc.arg('user_id');
