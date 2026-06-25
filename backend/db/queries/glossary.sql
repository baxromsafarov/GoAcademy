-- name: ListGlossaryTerms :many
SELECT * FROM glossary_terms
WHERE (sqlc.narg('language')::locale IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('q')::text IS NULL
       OR term ILIKE '%' || sqlc.narg('q') || '%'
       OR definition_markdown ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY term
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountGlossaryTerms :one
SELECT count(*) FROM glossary_terms
WHERE (sqlc.narg('language')::locale IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('q')::text IS NULL
       OR term ILIKE '%' || sqlc.narg('q') || '%'
       OR definition_markdown ILIKE '%' || sqlc.narg('q') || '%');
