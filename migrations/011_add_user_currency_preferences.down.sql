ALTER TABLE users
    DROP COLUMN IF EXISTS home_currency,
    DROP COLUMN IF EXISTS locale;
