-- name: CreateQuizAttempt :one
INSERT INTO quiz_attempts (user_id, quiz_id, score, passed, answers)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
