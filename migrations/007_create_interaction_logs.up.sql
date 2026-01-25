CREATE TABLE IF NOT EXISTS interaction_logs (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  user_input TEXT NOT NULL,
  system_prompt TEXT NOT NULL,
  ai_raw_response TEXT NOT NULL,
  bot_final_reply TEXT NOT NULL,
  duration_ms BIGINT NOT NULL,
  error TEXT,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE INDEX IF NOT EXISTS idx_interaction_logs_user ON interaction_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_interaction_logs_timestamp ON interaction_logs(timestamp);
