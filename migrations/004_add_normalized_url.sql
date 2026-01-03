-- +goose Up
ALTER TABLE entries
    ADD COLUMN normalized_url TEXT;

UPDATE entries
SET normalized_url = source_url
WHERE normalized_url IS NULL;

ALTER TABLE entries
    ALTER COLUMN normalized_url SET NOT NULL;

CREATE INDEX idx_entries_normalized_url ON entries(normalized_url);

-- +goose Down
DROP INDEX IF EXISTS idx_entries_normalized_url;

ALTER TABLE entries
    DROP COLUMN normalized_url;
