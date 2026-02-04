ALTER TABLE expenses ADD COLUMN source_message_id TEXT;
CREATE UNIQUE INDEX idx_expenses_source_message ON expenses(source_message_id) WHERE source_message_id IS NOT NULL;
