CREATE TABLE article_reads (
    user_id      uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_id   uuid        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    completed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, article_id)
);

CREATE INDEX idx_article_reads_user_id ON article_reads (user_id);
