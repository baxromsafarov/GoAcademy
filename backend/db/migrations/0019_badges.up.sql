-- badges are achievement definitions; user_badges records which user earned which.
CREATE TABLE badges (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    code            text        NOT NULL UNIQUE,
    title           text        NOT NULL,
    description     text        NOT NULL DEFAULT '',
    icon            text        NOT NULL DEFAULT '',
    criteria_type   text        NOT NULL,                 -- evaluated by the badge engine
    criteria_params jsonb       NOT NULL DEFAULT '{}',
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE user_badges (
    user_id    uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id   uuid        NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
    awarded_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, badge_id)
);

CREATE INDEX idx_user_badges_user ON user_badges (user_id, awarded_at);

-- Initial badge set. criteria_type/params are interpreted by gamification's engine:
--   activity_count_at_least {activity_type, count} | streak_at_least {days} | xp_at_least {xp}
INSERT INTO badges (code, title, description, icon, criteria_type, criteria_params) VALUES
    ('first_video',   'Первое видео',  'Просмотрено первое видео',   '▶️', 'activity_count_at_least', '{"activity_type":"video_completed","count":1}'),
    ('first_article', 'Первая статья', 'Прочитана первая статья',    '📖', 'activity_count_at_least', '{"activity_type":"article_read","count":1}'),
    ('first_quiz',    'Первый квиз',   'Пройден первый квиз',        '✅', 'activity_count_at_least', '{"activity_type":"quiz_passed","count":1}'),
    ('first_problem', 'Первая задача', 'Решена первая задача',       '🧩', 'activity_count_at_least', '{"activity_type":"problem_solved","count":1}'),
    ('streak_7',      'Неделя подряд', '7 дней активности подряд',   '🔥', 'streak_at_least',         '{"days":7}'),
    ('xp_100',        'Сотня XP',      'Набрано 100 XP',             '⭐', 'xp_at_least',             '{"xp":100}'),
    ('xp_500',        'Пятьсот XP',    'Набрано 500 XP',             '🌟', 'xp_at_least',             '{"xp":500}');
