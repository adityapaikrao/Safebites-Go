CREATE TABLE IF NOT EXISTS users (
    id                TEXT PRIMARY KEY,
    email             TEXT NOT NULL,
    name              TEXT,
    picture           TEXT,
    allergies         JSONB    NOT NULL DEFAULT '[]'::jsonb,
    diet_goals        JSONB    NOT NULL DEFAULT '[]'::jsonb,
    avoid_ingredients JSONB    NOT NULL DEFAULT '[]'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
