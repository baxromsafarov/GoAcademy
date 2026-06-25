CREATE TABLE mini_projects (
    id                   uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title                text        NOT NULL,
    description_markdown text        NOT NULL DEFAULT '',
    difficulty           difficulty  NOT NULL DEFAULT 'beginner',
    tags                 text[]      NOT NULL DEFAULT '{}',
    language             locale      NOT NULL DEFAULT 'en',
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE mini_project_steps (
    id         uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid    NOT NULL REFERENCES mini_projects(id) ON DELETE CASCADE,
    text       text    NOT NULL,
    position   integer NOT NULL DEFAULT 0
);

CREATE TABLE project_progress (
    user_id         uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id      uuid        NOT NULL REFERENCES mini_projects(id) ON DELETE CASCADE,
    completed_steps jsonb       NOT NULL DEFAULT '[]',     -- array of completed step ids
    updated_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, project_id)
);

CREATE TRIGGER mini_projects_set_updated_at
    BEFORE UPDATE ON mini_projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER project_progress_set_updated_at
    BEFORE UPDATE ON project_progress
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_mini_projects_difficulty     ON mini_projects (difficulty);
CREATE INDEX idx_mini_projects_language       ON mini_projects (language);
CREATE INDEX idx_mini_projects_created_at     ON mini_projects (created_at DESC);
CREATE INDEX idx_mini_projects_tags           ON mini_projects USING GIN (tags);
CREATE INDEX idx_mini_project_steps_project   ON mini_project_steps (project_id, position);
