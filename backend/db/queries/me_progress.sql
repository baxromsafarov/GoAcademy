-- name: ProgressSummary :one
-- Per-section completion counts for the authenticated user's dashboard.
-- Every column is alias-qualified: sqlc merges the scalar-subquery scopes, so an
-- unqualified user_id would read as ambiguous across the joined-in tables.
SELECT
    (SELECT count(*) FROM video_progress vp
        WHERE vp.user_id = sqlc.arg('user_id') AND vp.completed)                  AS videos_completed,
    (SELECT count(*) FROM article_reads ar
        WHERE ar.user_id = sqlc.arg('user_id'))                                  AS articles_read,
    (SELECT count(DISTINCT qa.quiz_id) FROM quiz_attempts qa
        WHERE qa.user_id = sqlc.arg('user_id') AND qa.passed)                    AS quizzes_passed,
    (SELECT count(DISTINCT ps.problem_id) FROM problem_submissions ps
        WHERE ps.user_id = sqlc.arg('user_id') AND ps.status = 'solved')         AS problems_solved,
    (SELECT count(*) FROM project_progress pp
        JOIN (
            SELECT mps.project_id, count(*) AS step_count
            FROM mini_project_steps mps GROUP BY mps.project_id
        ) s ON s.project_id = pp.project_id
        WHERE pp.user_id = sqlc.arg('user_id')
          AND jsonb_array_length(pp.completed_steps) >= s.step_count)            AS projects_completed;

-- name: ActivityHeatmap :many
-- Daily activity buckets in UTC (D-010) over the half-open range [from_ts, to_ts).
SELECT
    to_char((occurred_at AT TIME ZONE 'UTC')::date, 'YYYY-MM-DD') AS day,
    count(*)                            AS count,
    coalesce(sum(xp_earned), 0)::bigint AS xp
FROM activity_log
WHERE user_id = sqlc.arg('user_id')
  AND occurred_at >= sqlc.arg('from_ts')
  AND occurred_at <  sqlc.arg('to_ts')
GROUP BY day
ORDER BY day;
