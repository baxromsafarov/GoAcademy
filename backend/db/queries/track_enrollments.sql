-- name: EnrollTrack :exec
INSERT INTO track_enrollments (user_id, track_id)
VALUES ($1, $2)
ON CONFLICT (user_id, track_id) DO NOTHING;

-- name: UnenrollTrack :execrows
DELETE FROM track_enrollments WHERE user_id = $1 AND track_id = $2;

-- name: ListEnrolledTracks :many
SELECT t.* FROM tracks t
JOIN track_enrollments e ON e.track_id = t.id
WHERE e.user_id = $1
ORDER BY e.created_at DESC;

-- name: IsTrackEnrolled :one
SELECT EXISTS (
    SELECT 1 FROM track_enrollments WHERE user_id = $1 AND track_id = $2
);
