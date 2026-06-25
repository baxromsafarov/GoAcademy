-- name: IssueCertificate :one
-- Issues a certificate once per (user, track); a repeat conflicts and returns no
-- rows (the caller then fetches the existing one).
INSERT INTO certificates (user_id, track_id, certificate_code)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, track_id) DO NOTHING
RETURNING *;

-- name: GetUserTrackCertificate :one
SELECT * FROM certificates WHERE user_id = $1 AND track_id = $2;

-- name: ListUserCertificates :many
SELECT c.id, c.certificate_code, c.issued_at, c.track_id, t.title AS track_title
FROM certificates c
JOIN tracks t ON t.id = c.track_id
WHERE c.user_id = $1
ORDER BY c.issued_at DESC, c.id;

-- name: GetCertificateByCode :one
-- Public verification: who earned which track, and when.
SELECT c.certificate_code, c.issued_at, u.display_name, t.title AS track_title
FROM certificates c
JOIN users u ON u.id = c.user_id
JOIN tracks t ON t.id = c.track_id
WHERE c.certificate_code = $1;
