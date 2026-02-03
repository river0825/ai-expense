ALTER TABLE expenses
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS home_amount,
    DROP COLUMN IF EXISTS home_currency,
    DROP COLUMN IF EXISTS exchange_rate;

ALTER TABLE expenses
    RENAME COLUMN original_amount TO amount;
