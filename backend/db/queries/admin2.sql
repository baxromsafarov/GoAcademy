-- Admin write queries for CHAPTER 13.2 (tracks, cheatsheets, projects, glossary,
-- badges, daily challenges). Mounted behind RequireRole("admin").

-- tracks ----------------------------------------------------------------------
-- name: CreateTrack :one
INSERT INTO tracks (title, description, level, position, language)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateTrack :one
UPDATE tracks SET title = $2, description = $3, level = $4, position = $5, language = $6
WHERE id = $1
RETURNING *;

-- name: DeleteTrack :execrows
DELETE FROM tracks WHERE id = $1;

-- name: DeleteTrackItems :exec
DELETE FROM track_items WHERE track_id = $1;

-- name: CreateTrackItem :one
INSERT INTO track_items (track_id, content_type, content_id, position)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- cheatsheets -----------------------------------------------------------------
-- name: CreateCheatsheet :one
INSERT INTO cheatsheets (title, category, body_markdown, language)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateCheatsheet :one
UPDATE cheatsheets SET title = $2, category = $3, body_markdown = $4, language = $5
WHERE id = $1
RETURNING *;

-- name: DeleteCheatsheet :execrows
DELETE FROM cheatsheets WHERE id = $1;

-- mini projects ---------------------------------------------------------------
-- name: CreateProject :one
INSERT INTO mini_projects (title, description_markdown, difficulty, tags, language)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateProject :one
UPDATE mini_projects SET title = $2, description_markdown = $3, difficulty = $4, tags = $5, language = $6
WHERE id = $1
RETURNING *;

-- name: DeleteProject :execrows
DELETE FROM mini_projects WHERE id = $1;

-- name: DeleteProjectSteps :exec
DELETE FROM mini_project_steps WHERE project_id = $1;

-- name: CreateProjectStep :one
INSERT INTO mini_project_steps (project_id, text, position)
VALUES ($1, $2, $3)
RETURNING *;

-- glossary --------------------------------------------------------------------
-- name: CreateGlossaryTerm :one
INSERT INTO glossary_terms (term, definition_markdown, language)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateGlossaryTerm :one
UPDATE glossary_terms SET term = $2, definition_markdown = $3, language = $4
WHERE id = $1
RETURNING *;

-- name: DeleteGlossaryTerm :execrows
DELETE FROM glossary_terms WHERE id = $1;

-- badges ----------------------------------------------------------------------
-- name: CreateBadge :one
INSERT INTO badges (code, title, description, icon, criteria_type, criteria_params)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateBadge :one
UPDATE badges SET code = $2, title = $3, description = $4, icon = $5, criteria_type = $6, criteria_params = $7
WHERE id = $1
RETURNING *;

-- name: DeleteBadge :execrows
DELETE FROM badges WHERE id = $1;

-- daily challenges ------------------------------------------------------------
-- name: CreateDailyChallenge :one
INSERT INTO daily_challenges (challenge_date, content_type, content_id, bonus_xp)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateDailyChallenge :one
UPDATE daily_challenges SET challenge_date = $2, content_type = $3, content_id = $4, bonus_xp = $5
WHERE id = $1
RETURNING *;

-- name: DeleteDailyChallenge :execrows
DELETE FROM daily_challenges WHERE id = $1;
