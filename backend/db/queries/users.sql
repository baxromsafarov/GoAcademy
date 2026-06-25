-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name, locale)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: SetEmailVerified :exec
UPDATE users SET email_verified = true WHERE id = $1;

-- name: UpdatePasswordHash :exec
UPDATE users SET password_hash = $1 WHERE id = $2;

-- name: UpdateAvatarURL :one
UPDATE users SET avatar_url = $1 WHERE id = $2 RETURNING *;

-- name: UpdateUserProfile :one
-- Partial update: a NULL argument leaves the existing column value unchanged.
UPDATE users SET
    display_name = COALESCE(sqlc.narg('display_name'), display_name),
    bio          = COALESCE(sqlc.narg('bio'), bio),
    location     = COALESCE(sqlc.narg('location'), location),
    locale       = COALESCE(sqlc.narg('locale'), locale),
    is_public    = COALESCE(sqlc.narg('is_public'), is_public)
WHERE id = sqlc.arg('id')
RETURNING *;
