-- activity_log is the single, append-only journal of meaningful user actions.
-- It is the source of truth for the activity heatmap (CH10.2), the period
-- leaderboard (CH12) and XP (CH11). One row = one action.
CREATE TABLE activity_log (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type text        NOT NULL,                    -- e.g. 'video_completed', 'quiz_passed'
    ref_type      text        NOT NULL DEFAULT '',         -- e.g. 'video', 'article' ('' = ref-less)
    ref_id        uuid,                                    -- polymorphic, no FK (D-006); NULL = ref-less
    xp_earned     integer     NOT NULL DEFAULT 0 CHECK (xp_earned >= 0),
    occurred_at   timestamptz NOT NULL DEFAULT now()
);

-- Drives per-user, time-ranged reads (heatmap, recent activity, period leaderboard).
CREATE INDEX idx_activity_log_user_occurred ON activity_log (user_id, occurred_at DESC);
