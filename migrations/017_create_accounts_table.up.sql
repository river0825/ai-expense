-- Create accounts table with composite primary key (user_id, name)
CREATE TABLE IF NOT EXISTS accounts (
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, name),
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Create index for efficient lookups by user
CREATE INDEX IF NOT EXISTS idx_accounts_user ON accounts(user_id);

-- Seed accounts from existing distinct values in expenses table
-- This extracts all unique account names per user that are not empty
INSERT INTO accounts (user_id, name)
SELECT DISTINCT user_id, account
FROM expenses
WHERE account IS NOT NULL AND account != ''
ON CONFLICT (user_id, name) DO NOTHING;
