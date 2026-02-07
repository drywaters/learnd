-- +goose Up
ALTER TABLE entries ADD COLUMN tag TEXT;
UPDATE entries SET tag = tags[1];
ALTER TABLE entries DROP COLUMN tags;
CREATE INDEX idx_entries_tag ON entries(tag);

-- +goose Down
ALTER TABLE entries ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';
UPDATE entries SET tags = CASE WHEN tag IS NOT NULL THEN ARRAY[tag] ELSE '{}' END;
ALTER TABLE entries DROP COLUMN tag;
CREATE INDEX idx_entries_tags ON entries USING GIN(tags);
