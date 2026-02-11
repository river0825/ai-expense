CREATE TABLE IF NOT EXISTS conversation_states (
  user_id TEXT PRIMARY KEY,
  active_intent TEXT NOT NULL,
  pending_slots JSONB NOT NULL DEFAULT '{}'::jsonb,
  status TEXT NOT NULL DEFAULT 'pending',
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE INDEX IF NOT EXISTS idx_conversation_states_expires_at
  ON conversation_states(expires_at);
