## MODIFIED Requirements

### Requirement: Create Expense Record
The system SHALL create and persist a new expense record with date, amount, description, and category, and SHALL avoid creating records while an unresolved clarification for a non-expense intent is pending.

#### Scenario: Expense created successfully
- **WHEN** user sends "早餐$20"
- **THEN** system saves expense and responds "早餐 20元，已儲存"

#### Scenario: Expense with specific date
- **WHEN** user sends "昨天買水果$300"
- **THEN** system saves expense with date=yesterday and responds "水果 300元，已儲存"

#### Scenario: Multiple expenses in one message
- **WHEN** user sends "早餐$20午餐$30加油$200"
- **THEN** system saves all three expenses and responds with single message listing all three confirmations

#### Scenario: Incomplete expense (missing description)
- **WHEN** parser extracts "$100" without description
- **THEN** system asks user "這$100是什麼消費?" and does NOT create record until user clarifies

#### Scenario: Incomplete expense (missing amount)
- **WHEN** parser extracts "買菜" without amount
- **THEN** system asks user "買菜花了多少錢?" and does NOT create record until user clarifies

#### Scenario: Pending non-expense clarification blocks accidental expense creation
- **WHEN** system is waiting for currency answer from a previous currency-setting intent
- **AND** user replies with a supported currency code
- **THEN** system applies currency-setting action
- **AND** system does NOT create a new expense record from that reply
