-- name: ListProjects :many
SELECT * FROM mini_projects
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
  AND (sqlc.narg('q')::text              IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.arg('show_hidden')::bool OR NOT ('hidden' = ANY(tags)))
ORDER BY created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountProjects :one
SELECT count(*) FROM mini_projects
WHERE (sqlc.narg('difficulty')::difficulty IS NULL OR difficulty = sqlc.narg('difficulty'))
  AND (sqlc.narg('language')::locale     IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('tag')::text            IS NULL OR sqlc.narg('tag') = ANY(tags))
  AND (sqlc.narg('q')::text              IS NULL OR title ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.arg('show_hidden')::bool OR NOT ('hidden' = ANY(tags)));

-- name: GetProjectByID :one
SELECT * FROM mini_projects WHERE id = $1;

-- name: ListProjectSteps :many
SELECT * FROM mini_project_steps WHERE project_id = $1 ORDER BY position, id;

-- name: GetProjectProgress :one
SELECT * FROM project_progress WHERE user_id = $1 AND project_id = $2;

-- name: UpsertProjectProgress :one
INSERT INTO project_progress (user_id, project_id, completed_steps)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, project_id) DO UPDATE SET completed_steps = EXCLUDED.completed_steps
RETURNING *;
