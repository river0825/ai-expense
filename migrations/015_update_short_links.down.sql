DROP INDEX IF EXISTS idx_short_links_user_active;
ALTER TABLE short_links DROP COLUMN IF EXISTS deprecated_at;
ALTER TABLE short_links DROP COLUMN IF EXISTS user_id;
