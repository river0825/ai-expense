## ADDED Requirements

### Requirement: Create Expense Record
The system SHALL create and persist a new expense record with date, amount, description, and category.

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

### Requirement: Retrieve Expenses
The system SHALL retrieve expense records by date range, category, or all records.

#### Scenario: All expenses
- **WHEN** user requests "show all expenses" or "顯示所有消費"
- **THEN** system returns list of all expenses grouped by date

#### Scenario: Expenses by category
- **WHEN** user requests "food expenses this month" or "這個月的食物消費"
- **THEN** system returns only expenses in Food category for current month with total

#### Scenario: Expenses by date range
- **WHEN** user requests "expenses last week" or "上週的消費"
- **THEN** system returns all expenses from 7 days ago to today

#### Scenario: No expenses found
- **WHEN** user requests expenses for empty date range or category
- **THEN** system responds "沒有找到符合的消費"

### Requirement: Update Expense Record
The system SHALL allow users to update existing expense records.

#### Scenario: Update amount
- **WHEN** user identifies an expense by description and provides new amount
- **THEN** system updates the record and confirms "更新成功"

#### Scenario: Update category
- **WHEN** user reassigns an expense to different category
- **THEN** system updates record and confirms change

#### Scenario: Update date
- **WHEN** user corrects the date of an expense
- **THEN** system updates record and confirms change

### Requirement: Delete Expense Record
The system SHALL allow users to delete expense records.

#### Scenario: Delete by description
- **WHEN** user sends "delete 早餐" or "刪除 早餐"
- **THEN** system deletes the most recent matching expense and confirms "已刪除"

#### Scenario: Delete confirms before removal
- **WHEN** deletion could match multiple expenses
- **THEN** system asks user to confirm which record to delete
