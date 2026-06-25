-- daily_challenges: one challenge per calendar day (UTC, D-010), referencing a
-- piece of content polymorphically (D-006). Completing it grants bonus XP.
CREATE TABLE daily_challenges (
    id             uuid               PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_date date               NOT NULL UNIQUE,        -- unique index drives the by-date lookup
    content_type   track_content_type NOT NULL,
    content_id     uuid               NOT NULL,
    bonus_xp       integer            NOT NULL DEFAULT 0 CHECK (bonus_xp >= 0),
    created_at     timestamptz        NOT NULL DEFAULT now()
);

CREATE TABLE user_daily_challenges (
    user_id      uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id uuid        NOT NULL REFERENCES daily_challenges(id) ON DELETE CASCADE,
    completed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, challenge_id)
);
