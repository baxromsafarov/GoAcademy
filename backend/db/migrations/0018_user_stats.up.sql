-- user_stats is the denormalized per-user gamification aggregate: lifetime XP,
-- derived level, activity streak and the last active day. It is kept consistent
-- by writing it in the same transaction as the activity that drives it (CH11).
CREATE TABLE user_stats (
    user_id          uuid        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_xp         integer     NOT NULL DEFAULT 0 CHECK (total_xp >= 0),
    level            integer     NOT NULL DEFAULT 1 CHECK (level >= 1),
    current_streak   integer     NOT NULL DEFAULT 0 CHECK (current_streak >= 0),  -- driven in 11.2
    longest_streak   integer     NOT NULL DEFAULT 0 CHECK (longest_streak >= 0),  -- driven in 11.2
    last_active_date date,
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER user_stats_set_updated_at
    BEFORE UPDATE ON user_stats
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Leaderboard ranking (CH12) reads users ordered by XP.
CREATE INDEX idx_user_stats_total_xp ON user_stats (total_xp DESC);
