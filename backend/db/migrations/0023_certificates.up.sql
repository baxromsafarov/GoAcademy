-- certificates: proof a user completed a track. One per (user, track); the
-- certificate_code is a public, verifiable identifier.
CREATE TABLE certificates (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id         uuid        NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    certificate_code text        NOT NULL UNIQUE,
    issued_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, track_id)
);

CREATE INDEX idx_certificates_user ON certificates (user_id, issued_at DESC);
