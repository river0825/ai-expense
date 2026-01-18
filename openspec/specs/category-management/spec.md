# category-management Specification

## Purpose
TBD - created by archiving change add-conversational-expense-tracker. Update Purpose after archive.
## Requirements
### Requirement: Maintain Pre-defined Categories
The system SHALL provide a standard set of expense categories that users can leverage without creation.

#### Scenario: Default categories available
- **WHEN** system initializes
- **THEN** system provides these default categories: Food, Transport, Shopping, Entertainment, Other

#### Scenario: Retrieve category list
- **WHEN** user requests "show categories" or "顯示分類"
- **THEN** system returns list of all available categories (default + custom)

#### Scenario: Use default category
- **WHEN** user creates an expense and system suggests Food category
- **THEN** user can accept suggestion without adding custom category

### Requirement: Suggest Category from Expense Description (AI-Powered)
The system SHALL automatically suggest a category based on the expense description using Gemini 2.5 lite when the user creates a new expense, providing intelligent inference beyond simple keyword matching.

#### Scenario: Match description to category
- **WHEN** user sends "早餐$20"
- **THEN** system suggests category="Food" (matches "早餐")

#### Scenario: Transport matches
- **WHEN** user sends "加油$200"
- **THEN** system suggests category="Transport" (matches "加油")

#### Scenario: Shopping matches
- **WHEN** user sends "買衣服$500"
- **THEN** system suggests category="Shopping" (matches "衣服")

#### Scenario: Entertainment matches
- **WHEN** user sends "電影票$350"
- **THEN** system suggests category="Entertainment" (matches "電影")

#### Scenario: No clear match
- **WHEN** user sends "其他東西$100" (ambiguous description)
- **THEN** system suggests category="Other"

### Requirement: Allow User to Override Category Suggestion
The system SHALL allow users to accept or override the suggested category during expense creation.

#### Scenario: Accept suggested category
- **WHEN** user sends "早餐$20" and system suggests Food
- **THEN** user accepts and expense is saved with Food category

#### Scenario: Override with existing category
- **WHEN** user sends "早餐$20" and system suggests Food
- **AND** user responds with different category (e.g., "Entertainment")
- **THEN** expense is saved with Entertainment category

#### Scenario: Create new category on-the-fly
- **WHEN** user overrides suggestion and provides new category name (e.g., "Custom-Cafe")
- **THEN** system creates new category and saves expense with it

### Requirement: Create Custom Categories
The system SHALL allow users to add new categories beyond the defaults.

#### Scenario: Add new category
- **WHEN** user sends "add category 醫療" or "新增分類 醫療"
- **THEN** system creates category and confirms "分類 '醫療' 已新增"

#### Scenario: Custom category in future expenses
- **WHEN** user creates an expense "看醫生$1000"
- **THEN** system suggests "醫療" category (custom category created earlier)

#### Scenario: Duplicate category name
- **WHEN** user attempts to create category that already exists
- **THEN** system responds "分類已存在，請使用另一個名稱"

### Requirement: Manage Existing Categories
The system SHALL allow users to view, rename, or delete custom categories.

#### Scenario: List all categories
- **WHEN** user requests "show categories"
- **THEN** system returns all categories with counts of associated expenses

#### Scenario: Delete unused custom category
- **WHEN** user deletes a category with no associated expenses
- **THEN** system confirms deletion immediately

#### Scenario: Delete category with expenses
- **WHEN** user deletes a category with associated expenses
- **THEN** system asks user whether to reassign expenses to Other or keep category with new name

