-- +goose Up
CREATE TABLE IF NOT EXISTS urls (
    short_url    VARCHAR(10) PRIMARY KEY,
    original_url TEXT UNIQUE NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS urls;
