CREATE TABLE quiz_attempts (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id    uuid        NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    score      integer     NOT NULL,                 -- percent 0..100
    passed     boolean     NOT NULL,
    answers    jsonb       NOT NULL DEFAULT '{}',     -- {question_id: [option_id, ...]}
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_quiz_attempts_user_id ON quiz_attempts (user_id, created_at DESC);
CREATE INDEX idx_quiz_attempts_quiz_id ON quiz_attempts (quiz_id);
