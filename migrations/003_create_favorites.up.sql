CREATE TABLE IF NOT EXISTS favorites (
    id            SERIAL      PRIMARY KEY,
    user_id       TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_name  TEXT        NOT NULL,
    brand         TEXT        NOT NULL DEFAULT '',
    safety_score  INTEGER,
    image         TEXT,
    added_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, product_name)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
