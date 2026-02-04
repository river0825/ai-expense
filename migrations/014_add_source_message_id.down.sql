DROP INDEX IF EXISTS idx_expenses_source_message;
ALTER TABLE expenses DROP COLUMN IF EXISTS source_message_id;
