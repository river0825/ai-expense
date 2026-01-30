ALTER TABLE users
    DROP COLUMN IF EXISTS home_currency;

ALTER TABLE users
    DROP COLUMN IF EXISTS locale;
