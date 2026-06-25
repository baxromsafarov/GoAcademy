-- bookmarks: a user marking content to revisit. Polymorphic (content_type,
-- content_id) — D-006. At most one bookmark per (user, content): no duplicates.
CREATE TABLE bookmarks (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content_type text        NOT NULL,
    content_id   uuid        NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, content_type, content_id)
);

CREATE INDEX idx_bookmarks_user ON bookmarks (user_id, created_at DESC);
