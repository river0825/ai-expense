ALTER TABLE users
    ADD COLUMN IF NOT EXISTS home_currency TEXT NOT NULL DEFAULT 'TWD',
    ADD COLUMN IF NOT EXISTS locale TEXT NOT NULL DEFAULT 'zh-TW';

UPDATE users
SET home_currency = 'TWD'
WHERE home_currency IS NULL;

UPDATE users
SET locale = 'zh-TW'
WHERE locale IS NULL;
