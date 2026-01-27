-- Add cost_note column to ai_cost_logs table for audit trail
ALTER TABLE ai_cost_logs ADD COLUMN cost_note TEXT;
