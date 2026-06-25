-- name: LeaderboardAllTime :many
-- Public, non-blocked users ranked by lifetime XP.
SELECT u.id, u.display_name, u.avatar_url, us.total_xp::bigint AS xp
FROM user_stats us
JOIN users u ON u.id = us.user_id
WHERE u.is_public AND NOT u.is_blocked
ORDER BY us.total_xp DESC, u.id
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: LeaderboardPeriod :many
-- Public, non-blocked users ranked by XP earned in the half-open window [from, to).
SELECT u.id, u.display_name, u.avatar_url, COALESCE(SUM(al.xp_earned), 0)::bigint AS xp
FROM users u
JOIN activity_log al ON al.user_id = u.id
WHERE u.is_public AND NOT u.is_blocked
  AND al.occurred_at >= sqlc.arg('from_ts')
  AND al.occurred_at <  sqlc.arg('to_ts')
GROUP BY u.id, u.display_name, u.avatar_url
HAVING COALESCE(SUM(al.xp_earned), 0) > 0
ORDER BY xp DESC, u.id
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');
