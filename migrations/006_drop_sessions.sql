-- +goose Up
DROP TABLE IF EXISTS sessions;

-- +goose Down
CREATE TABLE sessions (
    token      TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

