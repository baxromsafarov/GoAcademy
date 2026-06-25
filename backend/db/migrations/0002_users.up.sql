-- Roles and supported UI locales as enum types.
CREATE TYPE user_role AS ENUM ('student', 'admin');
CREATE TYPE locale AS ENUM ('ru', 'en', 'uz', 'ja');

CREATE TABLE users (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    email           citext      NOT NULL UNIQUE,           -- case-insensitive uniqueness
    password_hash   text        NOT NULL,                  -- argon2id PHC string
    email_verified  boolean     NOT NULL DEFAULT false,
    role            user_role   NOT NULL DEFAULT 'student',
    is_blocked      boolean     NOT NULL DEFAULT false,
    display_name    text        NOT NULL,
    avatar_url      text,                                  -- nullable: no avatar by default
    bio             text        NOT NULL DEFAULT '',
    location        text        NOT NULL DEFAULT '',
    locale          locale      NOT NULL DEFAULT 'en',
    is_public       boolean     NOT NULL DEFAULT true,     -- visible on leaderboard
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- Keep updated_at current on every UPDATE (helper from migration 0001).
CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
