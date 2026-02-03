## ADDED Requirements
### Requirement: Multi-currency Expense Storage
The system SHALL capture the original currency and amount for every parsed expense, convert it into the users configured home currency using the most recent available exchange rate, and persist both values with the applied rate for future reporting.

#### Scenario: Expense with explicit foreign currency
- **WHEN** user sends "早餐 300 日幣"
- **THEN** the system saves the expense with `currency="JPY"`, `original_amount=300`, `home_currency` equal to the user preference (e.g., `"TWD"`), `home_amount` computed via the cached rate, and `exchange_rate` stored for auditing
- **AND** the confirmation message displays `"¥300 (≈NT$64)"`

#### Scenario: Expense without currency uses home default
- **WHEN** user sends "午餐 150"
- **AND** the message omits a currency indicator
- **THEN** the system records both `currency` and `home_currency` as the users configured home currency and sets `home_amount` equal to `original_amount`

#### Scenario: Unsupported/ambiguous currency falls back
- **WHEN** user sends "$20 咖啡"
- **AND** the symbol is ambiguous
- **THEN** the system defaults to the users home currency unless the user has an active explicit context (future trip modes) and stores `currency_original="$"` for auditing

#### Scenario: Exchange rate cache missing target day
- **WHEN** user records an expense on a weekend without a same-day exchange rate
- **THEN** the system uses the most recent available business-day rate before the expense date, stores that rate date, and surfaces it via APIs/reports
