CREATE TABLE IF NOT EXISTS scans (
    id            TEXT        PRIMARY KEY,
    user_id       TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_name  TEXT        NOT NULL,
    brand         TEXT        NOT NULL DEFAULT '',
    image         TEXT,
    safety_score  INTEGER     NOT NULL,
    is_safe       BOOLEAN     NOT NULL,
    ingredients   JSONB       NOT NULL DEFAULT '[]'::jsonb,
    timestamp     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scans_user_id   ON scans(user_id);
CREATE INDEX IF NOT EXISTS idx_scans_timestamp ON scans(timestamp DESC);
