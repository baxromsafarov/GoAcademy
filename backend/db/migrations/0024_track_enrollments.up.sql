-- track_enrollments: a student opting into a learning track so it surfaces on
-- their dashboard. One row per (user, track); removing either side cascades.
CREATE TABLE track_enrollments (
    user_id    uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id   uuid        NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, track_id)
);

CREATE INDEX idx_track_enrollments_user ON track_enrollments (user_id, created_at DESC);
