ALTER TABLE expenses
    RENAME COLUMN amount TO original_amount;

ALTER TABLE expenses
    ADD COLUMN currency TEXT;

ALTER TABLE expenses
    ADD COLUMN home_amount DECIMAL;

ALTER TABLE expenses
    ADD COLUMN home_currency TEXT;

ALTER TABLE expenses
    ADD COLUMN exchange_rate NUMERIC;

UPDATE expenses
SET currency = 'TWD',
    home_currency = 'TWD',
    home_amount = original_amount,
    exchange_rate = 1.0
WHERE currency IS NULL;

ALTER TABLE expenses
    ALTER COLUMN currency SET NOT NULL;

ALTER TABLE expenses
    ALTER COLUMN currency SET DEFAULT 'TWD';

ALTER TABLE expenses
    ALTER COLUMN home_currency SET NOT NULL;

ALTER TABLE expenses
    ALTER COLUMN home_currency SET DEFAULT 'TWD';

ALTER TABLE expenses
    ALTER COLUMN home_amount SET NOT NULL;

ALTER TABLE expenses
    ALTER COLUMN exchange_rate SET NOT NULL;

ALTER TABLE expenses
    ALTER COLUMN exchange_rate SET DEFAULT 1.0;
