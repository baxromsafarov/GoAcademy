-- Difficulty is shared by content types (videos, articles, problems, ...).
CREATE TYPE difficulty AS ENUM ('beginner', 'intermediate', 'advanced');

CREATE TABLE videos (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title            text        NOT NULL,
    description      text        NOT NULL DEFAULT '',
    youtube_id       text        NOT NULL,
    duration_seconds integer     NOT NULL DEFAULT 0,
    difficulty       difficulty  NOT NULL DEFAULT 'beginner',
    tags             text[]      NOT NULL DEFAULT '{}',
    language         locale      NOT NULL DEFAULT 'en',     -- content language (D-008)
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER videos_set_updated_at
    BEFORE UPDATE ON videos
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_videos_difficulty ON videos (difficulty);
CREATE INDEX idx_videos_language   ON videos (language);
CREATE INDEX idx_videos_created_at ON videos (created_at DESC);
CREATE INDEX idx_videos_tags       ON videos USING GIN (tags);
