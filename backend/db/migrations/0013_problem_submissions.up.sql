CREATE TYPE submission_status AS ENUM ('attempted', 'solved');

CREATE TABLE problem_submissions (
    id         uuid              PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid              NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id uuid              NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    status     submission_status NOT NULL DEFAULT 'attempted',
    code       text              NOT NULL DEFAULT '',
    language   text              NOT NULL DEFAULT 'go',
    verdict    jsonb,            -- NULL in MVP; populated by the online judge (CHAPTER 17)
    created_at timestamptz       NOT NULL DEFAULT now()
);

CREATE INDEX idx_problem_submissions_user_id      ON problem_submissions (user_id, created_at DESC);
CREATE INDEX idx_problem_submissions_problem_id   ON problem_submissions (problem_id);
CREATE INDEX idx_problem_submissions_user_problem ON problem_submissions (user_id, problem_id);
