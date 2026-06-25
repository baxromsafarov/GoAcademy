CREATE TYPE track_content_type AS ENUM ('video', 'article', 'quiz', 'problem', 'project');

CREATE TABLE tracks (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title       text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    level       difficulty  NOT NULL DEFAULT 'beginner',
    position    integer     NOT NULL DEFAULT 0,
    language    locale      NOT NULL DEFAULT 'en',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Polymorphic link to content (content_type + content_id), no FK by design (D-006).
CREATE TABLE track_items (
    id           uuid               PRIMARY KEY DEFAULT gen_random_uuid(),
    track_id     uuid               NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    content_type track_content_type NOT NULL,
    content_id   uuid               NOT NULL,
    position     integer            NOT NULL DEFAULT 0
);

CREATE TRIGGER tracks_set_updated_at
    BEFORE UPDATE ON tracks
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_tracks_position ON tracks (position, created_at DESC);
CREATE INDEX idx_tracks_language ON tracks (language);
CREATE UNIQUE INDEX idx_track_items_unique  ON track_items (track_id, content_type, content_id);
CREATE INDEX idx_track_items_track_position ON track_items (track_id, position);
