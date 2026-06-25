CREATE TABLE email_verification_tokens (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  text        NOT NULL,                  -- SHA-256 of the token; plaintext is emailed, never stored
    expires_at  timestamptz NOT NULL,
    used_at     timestamptz,                           -- set when the token is consumed
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_email_verification_tokens_token_hash ON email_verification_tokens (token_hash);
CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);
