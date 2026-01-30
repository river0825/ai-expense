ALTER TABLE expenses
    DROP COLUMN IF EXISTS currency;

ALTER TABLE expenses
    DROP COLUMN IF EXISTS home_amount;

ALTER TABLE expenses
    DROP COLUMN IF EXISTS home_currency;

ALTER TABLE expenses
    DROP COLUMN IF EXISTS exchange_rate;

ALTER TABLE expenses
    RENAME COLUMN original_amount TO amount;
