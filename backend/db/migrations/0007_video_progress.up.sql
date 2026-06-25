CREATE TABLE video_progress (
    user_id               uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id              uuid        NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    watched_percent       integer     NOT NULL DEFAULT 0,
    last_position_seconds integer     NOT NULL DEFAULT 0,
    completed             boolean     NOT NULL DEFAULT false,
    updated_at            timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, video_id)
);

CREATE TRIGGER video_progress_set_updated_at
    BEFORE UPDATE ON video_progress
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_video_progress_user_id ON video_progress (user_id);
