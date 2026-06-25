CREATE TABLE articles (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title         text        NOT NULL,
    slug          citext      NOT NULL UNIQUE,           -- case-insensitive unique slug
    body_markdown text        NOT NULL DEFAULT '',
    difficulty    difficulty  NOT NULL DEFAULT 'beginner',
    tags          text[]      NOT NULL DEFAULT '{}',
    language      locale      NOT NULL DEFAULT 'en',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER articles_set_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_articles_difficulty ON articles (difficulty);
CREATE INDEX idx_articles_language   ON articles (language);
CREATE INDEX idx_articles_created_at ON articles (created_at DESC);
CREATE INDEX idx_articles_tags       ON articles USING GIN (tags);
