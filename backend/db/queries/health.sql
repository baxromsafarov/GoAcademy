-- name: Now :one
-- Demo query proving the sqlc pipeline end-to-end; also usable as a DB liveness probe.
SELECT now()::timestamptz AS now;
