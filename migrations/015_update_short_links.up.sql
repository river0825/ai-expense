-- Add user_id column to short_links for tracking ownership
ALTER TABLE short_links ADD COLUMN user_id TEXT;

-- Add deprecated flag (NULL = active, timestamp = when deprecated)
ALTER TABLE short_links ADD COLUMN deprecated_at TIMESTAMP;

-- Create index for finding active links by user
CREATE INDEX idx_short_links_user_active ON short_links(user_id) WHERE deprecated_at IS NULL;
