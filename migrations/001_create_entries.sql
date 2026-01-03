-- +goose Up
CREATE TABLE entries (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- User input
    source_url           TEXT NOT NULL,
    normalized_url       TEXT NOT NULL,
    tags                 TEXT[] NOT NULL DEFAULT '{}',
    time_spent_seconds   INTEGER,
    quantity             INTEGER,
    notes                TEXT,

    -- Enriched fields
    canonical_url        TEXT,
    domain               TEXT,
    source_type          TEXT NOT NULL DEFAULT 'other',
    title                TEXT,
    description          TEXT,
    published_at         TIMESTAMPTZ,
    runtime_seconds      INTEGER,
    metadata_json        JSONB,

    -- Enrichment status
    enrichment_status    TEXT NOT NULL DEFAULT 'pending',
    enrichment_error     TEXT,
    enriched_at          TIMESTAMPTZ,

    -- Summary fields
    summary_text         TEXT,
    summary_status       TEXT NOT NULL DEFAULT 'pending',
    summary_error        TEXT,
    summary_provider     TEXT,
    summary_model        TEXT,
    summary_version      TEXT,
    summary_generated_at TIMESTAMPTZ
);

-- Indexes for common queries
CREATE INDEX idx_entries_created_at ON entries(created_at DESC);
CREATE INDEX idx_entries_source_type ON entries(source_type);
CREATE INDEX idx_entries_tags ON entries USING GIN(tags);
CREATE INDEX idx_entries_domain ON entries(domain);
CREATE INDEX idx_entries_normalized_url ON entries(normalized_url);

-- Partial indexes for pending work
CREATE INDEX idx_entries_enrichment_pending ON entries(enrichment_status)
    WHERE enrichment_status IN ('pending', 'processing');
CREATE INDEX idx_entries_summary_pending ON entries(summary_status)
    WHERE summary_status IN ('pending', 'processing');

-- Function to auto-update updated_at column
DROP FUNCTION IF EXISTS update_updated_at() CASCADE;

-- +goose StatementBegin
CREATE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Trigger to call update_updated_at before each row update
DROP TRIGGER IF EXISTS trg_entries_updated_at ON entries;
CREATE TRIGGER trg_entries_updated_at
    BEFORE UPDATE ON entries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_entries_updated_at ON entries;
DROP FUNCTION IF EXISTS update_updated_at();
DROP TABLE entries;
