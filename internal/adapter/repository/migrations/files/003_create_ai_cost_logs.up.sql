CREATE TABLE IF NOT EXISTS ai_cost_logs (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  operation TEXT NOT NULL,
  provider TEXT NOT NULL,
  model TEXT NOT NULL,
  input_tokens INT DEFAULT 0,
  output_tokens INT DEFAULT 0,
  total_tokens INT DEFAULT 0,
  cost DECIMAL NOT NULL,
  currency TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE INDEX IF NOT EXISTS idx_ai_cost_logs_user ON ai_cost_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_cost_logs_created_at ON ai_cost_logs(created_at);
