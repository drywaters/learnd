-- +goose Up
CREATE INDEX idx_entries_normalized_url_created_at ON entries(normalized_url, created_at DESC);
CREATE INDEX idx_entries_enrichment_pending_created_at ON entries(enrichment_status, created_at)
    WHERE enrichment_status IN ('pending', 'processing');
CREATE INDEX idx_entries_summary_pending_created_at ON entries(summary_status, enrichment_status, created_at)
    WHERE summary_status IN ('pending', 'processing');

DROP INDEX IF EXISTS idx_entries_normalized_url;
DROP INDEX IF EXISTS idx_entries_enrichment_pending;
DROP INDEX IF EXISTS idx_entries_summary_pending;

-- +goose Down
CREATE INDEX idx_entries_normalized_url ON entries(normalized_url);
CREATE INDEX idx_entries_enrichment_pending ON entries(enrichment_status)
    WHERE enrichment_status IN ('pending', 'processing');
CREATE INDEX idx_entries_summary_pending ON entries(summary_status)
    WHERE summary_status IN ('pending', 'processing');

DROP INDEX IF EXISTS idx_entries_normalized_url_created_at;
DROP INDEX IF EXISTS idx_entries_enrichment_pending_created_at;
DROP INDEX IF EXISTS idx_entries_summary_pending_created_at;
