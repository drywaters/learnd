-- +goose Up
CREATE TABLE summary_cache (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url_hash      TEXT NOT NULL UNIQUE,
    canonical_url TEXT NOT NULL,
    summary_text  TEXT NOT NULL,
    provider      TEXT NOT NULL,
    model         TEXT NOT NULL,
    version       TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_summary_cache_url_hash ON summary_cache(url_hash);

-- +goose Down
DROP TABLE summary_cache;
