CREATE TABLE problems (
    id                          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title                       text        NOT NULL,
    slug                        citext      NOT NULL UNIQUE,
    statement_markdown          text        NOT NULL DEFAULT '',
    difficulty                  difficulty  NOT NULL DEFAULT 'beginner',
    reference_solution_markdown text        NOT NULL DEFAULT '',     -- hidden until the user marks the problem solved
    sample_io                   jsonb       NOT NULL DEFAULT '[]',   -- [{input, output}, ...] shown in the statement
    tags                        text[]      NOT NULL DEFAULT '{}',
    language                    locale      NOT NULL DEFAULT 'en',
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now()
);

-- Full test set for the online judge (CHAPTER 17). is_sample marks cases that may
-- be shown to students; the rest are hidden.
CREATE TABLE problem_test_cases (
    id              uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id      uuid    NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    input           text    NOT NULL DEFAULT '',
    expected_output text    NOT NULL DEFAULT '',
    is_sample       boolean NOT NULL DEFAULT false,
    position        integer NOT NULL DEFAULT 0
);

CREATE TRIGGER problems_set_updated_at
    BEFORE UPDATE ON problems
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_problems_difficulty            ON problems (difficulty);
CREATE INDEX idx_problems_language              ON problems (language);
CREATE INDEX idx_problems_created_at            ON problems (created_at DESC);
CREATE INDEX idx_problems_tags                  ON problems USING GIN (tags);
CREATE INDEX idx_problem_test_cases_problem_id  ON problem_test_cases (problem_id, position);
