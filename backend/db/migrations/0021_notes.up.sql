-- notes: a user's private annotation attached polymorphically to any content
-- (content_type, content_id) — D-006. Visible only to its owner.
CREATE TABLE notes (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content_type text        NOT NULL,
    content_id   uuid        NOT NULL,
    body         text        NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER notes_set_updated_at
    BEFORE UPDATE ON notes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_notes_user    ON notes (user_id, created_at DESC);
CREATE INDEX idx_notes_content ON notes (content_type, content_id);
