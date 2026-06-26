-- name: ListRecentCompletions :many
-- A unified, most-recent-first feed of content the user has finished: videos
-- watched, articles read, quizzes passed and problems solved, with titles.
SELECT content_type, content_id, title, completed_at
FROM (
    SELECT 'video'::text AS content_type, v.id AS content_id, v.title AS title,
           vp.updated_at AS completed_at
    FROM video_progress vp
    JOIN videos v ON v.id = vp.video_id
    WHERE vp.user_id = $1 AND vp.completed

    UNION ALL

    SELECT 'article', a.id, a.title, ar.completed_at
    FROM article_reads ar
    JOIN articles a ON a.id = ar.article_id
    WHERE ar.user_id = $1

    UNION ALL

    SELECT 'quiz', q.id, q.title, max(qa.created_at)
    FROM quiz_attempts qa
    JOIN quizzes q ON q.id = qa.quiz_id
    WHERE qa.user_id = $1 AND qa.passed
    GROUP BY q.id, q.title

    UNION ALL

    SELECT 'problem', p.id, p.title, max(ps.created_at)
    FROM problem_submissions ps
    JOIN problems p ON p.id = ps.problem_id
    WHERE ps.user_id = $1 AND ps.status = 'solved'
    GROUP BY p.id, p.title
) c
ORDER BY completed_at DESC
LIMIT $2;
