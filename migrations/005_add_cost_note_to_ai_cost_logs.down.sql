-- Rollback: Remove cost_note column from ai_cost_logs
ALTER TABLE ai_cost_logs DROP COLUMN cost_note;
