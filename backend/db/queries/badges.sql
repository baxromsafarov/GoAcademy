-- name: ListUnearnedBadges :many
-- Badges the user has not yet earned (the engine only evaluates these).
SELECT * FROM badges b
WHERE NOT EXISTS (
    SELECT 1 FROM user_badges ub WHERE ub.badge_id = b.id AND ub.user_id = $1
)
ORDER BY b.code;

-- name: AwardBadge :exec
INSERT INTO user_badges (user_id, badge_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: ActivityRefCount :one
-- Number of distinct referenced entities a user has for an activity type
-- (e.g. how many distinct problems were solved).
SELECT count(DISTINCT ref_id) FROM activity_log
WHERE user_id = sqlc.arg('user_id')
  AND activity_type = sqlc.arg('activity_type')
  AND ref_id IS NOT NULL;

-- name: ListUserBadges :many
SELECT b.code, b.title, b.description, b.icon, ub.awarded_at
FROM user_badges ub
JOIN badges b ON b.id = ub.badge_id
WHERE ub.user_id = $1
ORDER BY ub.awarded_at, b.code;
