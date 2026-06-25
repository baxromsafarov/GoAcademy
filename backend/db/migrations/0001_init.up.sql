-- Case-insensitive text type, used for unique email storage.
CREATE EXTENSION IF NOT EXISTS citext;

-- Trigger helper: any table with an updated_at column attaches a BEFORE UPDATE
-- trigger calling this function to keep updated_at current automatically.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
