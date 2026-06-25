-- name: GetDailyChallengeByDate :one
SELECT * FROM daily_challenges WHERE challenge_date = $1;

-- name: IsDailyChallengeCompleted :one
SELECT EXISTS(
    SELECT 1 FROM user_daily_challenges
    WHERE user_id = sqlc.arg('user_id') AND challenge_id = sqlc.arg('challenge_id')
);

-- name: CompleteDailyChallenge :one
-- Marks the challenge done for the user. Returns the completion time the first
-- time; a repeat conflicts and returns no rows (idempotent — no double reward).
INSERT INTO user_daily_challenges (user_id, challenge_id)
VALUES (sqlc.arg('user_id'), sqlc.arg('challenge_id'))
ON CONFLICT DO NOTHING
RETURNING completed_at;
