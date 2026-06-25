-- Admin write queries (mutations behind RequireRole("admin")).

-- videos ----------------------------------------------------------------------
-- name: CreateVideo :one
INSERT INTO videos (title, description, youtube_id, duration_seconds, difficulty, tags, language)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateVideo :one
UPDATE videos SET
    title = $2, description = $3, youtube_id = $4, duration_seconds = $5,
    difficulty = $6, tags = $7, language = $8
WHERE id = $1
RETURNING *;

-- name: DeleteVideo :execrows
DELETE FROM videos WHERE id = $1;

-- articles --------------------------------------------------------------------
-- name: CreateArticle :one
INSERT INTO articles (title, slug, body_markdown, difficulty, tags, language)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateArticle :one
UPDATE articles SET
    title = $2, slug = $3, body_markdown = $4, difficulty = $5, tags = $6, language = $7
WHERE id = $1
RETURNING *;

-- name: DeleteArticle :execrows
DELETE FROM articles WHERE id = $1;

-- quizzes ---------------------------------------------------------------------
-- name: CreateQuiz :one
INSERT INTO quizzes (title, description, pass_threshold, difficulty, tags, language)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateQuiz :one
UPDATE quizzes SET
    title = $2, description = $3, pass_threshold = $4, difficulty = $5, tags = $6, language = $7
WHERE id = $1
RETURNING *;

-- name: DeleteQuiz :execrows
DELETE FROM quizzes WHERE id = $1;

-- name: DeleteQuizQuestions :exec
DELETE FROM quiz_questions WHERE quiz_id = $1;

-- name: CreateQuizQuestion :one
INSERT INTO quiz_questions (quiz_id, prompt, type, position)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateQuizOption :one
INSERT INTO quiz_options (question_id, text, is_correct, position)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- problems --------------------------------------------------------------------
-- name: CreateProblem :one
INSERT INTO problems (title, slug, statement_markdown, difficulty, reference_solution_markdown, sample_io, tags, language)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateProblem :one
UPDATE problems SET
    title = $2, slug = $3, statement_markdown = $4, difficulty = $5,
    reference_solution_markdown = $6, sample_io = $7, tags = $8, language = $9
WHERE id = $1
RETURNING *;

-- name: DeleteProblem :execrows
DELETE FROM problems WHERE id = $1;

-- name: DeleteProblemTestCases :exec
DELETE FROM problem_test_cases WHERE problem_id = $1;

-- name: CreateProblemTestCase :one
INSERT INTO problem_test_cases (problem_id, input, expected_output, is_sample, position)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
