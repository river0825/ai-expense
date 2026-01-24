-- Create policies table
CREATE TABLE IF NOT EXISTS policies (
  id TEXT PRIMARY KEY,
  key TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  version TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial Privacy Policy (using ON CONFLICT for PostgreSQL/SQLite compatibility)
INSERT INTO policies (id, key, title, content, version, created_at, updated_at)
VALUES (
  'policy_privacy',
  'privacy_policy',
  'Privacy Policy',
  '# Privacy Policy\n\nLast updated: January 2024\n\n## 1. Introduction\nWe respect your privacy and are committed to protecting your personal data.\n\n## 2. Data We Collect\nWe collect data you provide directly to us, such as expense details and messages sent to the bot.\n\n## 3. How We Use Your Data\nWe use your data to provide and improve the expense tracking service.\n\n## 4. Data Sharing\nWe do not share your personal data with third parties except as necessary to provide the service (e.g., AI providers).',
  '1.0',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT (key) DO NOTHING;

-- Seed initial Terms of Use
INSERT INTO policies (id, key, title, content, version, created_at, updated_at)
VALUES (
  'policy_terms',
  'terms_of_use',
  'Terms of Use',
  '# Terms of Use\n\nLast updated: January 2024\n\n## 1. Acceptance of Terms\nBy using this service, you agree to be bound by these Terms.\n\n## 2. Use of Service\nYou agree to use the service only for lawful purposes.\n\n## 3. Disclaimer\nThe service is provided "as is" without warranties of any kind.\n\n## 4. Limitation of Liability\nWe shall not be liable for any indirect, incidental, or consequential damages.',
  '1.0',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT (key) DO NOTHING;
