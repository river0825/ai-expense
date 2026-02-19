CREATE INDEX IF NOT EXISTS idx_expenses_source_message_pattern ON expenses (source_message_id text_pattern_ops);
