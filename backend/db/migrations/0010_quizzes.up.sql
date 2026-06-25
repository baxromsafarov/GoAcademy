CREATE TYPE quiz_question_type AS ENUM ('single', 'multiple');

CREATE TABLE quizzes (
    id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title          text        NOT NULL,
    description    text        NOT NULL DEFAULT '',
    pass_threshold integer     NOT NULL DEFAULT 70,        -- percent required to pass
    difficulty     difficulty  NOT NULL DEFAULT 'beginner',
    tags           text[]      NOT NULL DEFAULT '{}',
    language       locale      NOT NULL DEFAULT 'en',
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE quiz_questions (
    id       uuid               PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id  uuid               NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    prompt   text               NOT NULL,
    type     quiz_question_type NOT NULL DEFAULT 'single',
    position integer            NOT NULL DEFAULT 0
);

CREATE TABLE quiz_options (
    id          uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id uuid    NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    text        text    NOT NULL,
    is_correct  boolean NOT NULL DEFAULT false,            -- never exposed to students
    position    integer NOT NULL DEFAULT 0
);

CREATE TRIGGER quizzes_set_updated_at
    BEFORE UPDATE ON quizzes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_quizzes_difficulty       ON quizzes (difficulty);
CREATE INDEX idx_quizzes_language         ON quizzes (language);
CREATE INDEX idx_quizzes_created_at       ON quizzes (created_at DESC);
CREATE INDEX idx_quizzes_tags             ON quizzes USING GIN (tags);
CREATE INDEX idx_quiz_questions_quiz_id   ON quiz_questions (quiz_id, position);
CREATE INDEX idx_quiz_options_question_id ON quiz_options (question_id, position);
