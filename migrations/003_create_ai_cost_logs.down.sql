-- Drop ai_cost_logs indexes
DROP INDEX IF EXISTS idx_ai_cost_logs_created_at;
DROP INDEX IF EXISTS idx_ai_cost_logs_user;

-- Drop ai_cost_logs table
DROP TABLE IF EXISTS ai_cost_logs;
