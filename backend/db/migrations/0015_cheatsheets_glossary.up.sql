CREATE TABLE cheatsheets (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title         text        NOT NULL,
    category      text        NOT NULL DEFAULT '',
    body_markdown text        NOT NULL DEFAULT '',
    language      locale      NOT NULL DEFAULT 'en',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE glossary_terms (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    term                citext      NOT NULL UNIQUE,        -- case-insensitive unique term
    definition_markdown text        NOT NULL DEFAULT '',
    language            locale      NOT NULL DEFAULT 'en',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER cheatsheets_set_updated_at
    BEFORE UPDATE ON cheatsheets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER glossary_terms_set_updated_at
    BEFORE UPDATE ON glossary_terms
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_cheatsheets_category ON cheatsheets (category);
CREATE INDEX idx_cheatsheets_title    ON cheatsheets (title);
CREATE INDEX idx_glossary_terms_term  ON glossary_terms (term);
