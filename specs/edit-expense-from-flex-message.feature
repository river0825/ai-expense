# Feature: Edit Expense from LINE Flex Message

## Overview
Users should be able to edit their recorded expenses directly from the LINE confirmation message by clicking an "Edit" button that opens the dashboard with the expense in edit mode.

## User Story
As a user who just recorded an expense via LINE,
I want to see an "Edit" button for each expense in the confirmation message,
So that I can quickly correct any mistakes without searching for the expense in the dashboard.

## Scenarios

### Scenario 1: Flex message displays edit button for each expense
**Status**: [-] In progress - Tests written

Given I have recorded 2 expenses via LINE chat:
  - "Lunch bento $85"
  - "Coffee $50"
When I receive the expense confirmation flex message
Then I should see an "Edit" button for the "Lunch bento" expense
And I should see an "Edit" button for the "Coffee" expense

### Scenario 2: Edit button contains correct deep link URL
**Status**: [-] In progress - Tests written

Given I have recorded an expense "Lunch $85" with ID "exp_123"
When the flex message is generated
Then the edit button should contain a URI action
And the URI should be "https://{dashboard_url}/dashboard/expenses?edit=exp_123"

### Scenario 3: Clicking edit button opens dashboard with expense in edit mode
**Status**: [ ] Not started

Given I have recorded an expense "Lunch $85" with ID "exp_123"
And I receive the flex message with an edit button
When I click the "Edit" button
Then LINE should open the dashboard URL in browser
And the dashboard should load with "?edit=exp_123" query parameter
And the expense "Lunch $85" should be automatically opened in edit mode

### Scenario 4: Edit button supports multiple languages
**Status**: [-] In progress - Tests written

Given my LINE language is set to "zh-TW"
When I receive an expense confirmation flex message
Then the edit button text should be "編輯"

Given my LINE language is set to "en"
When I receive an expense confirmation flex message
Then the edit button text should be "Edit"

Given my LINE language is set to "ja"
When I receive an expense confirmation flex message
Then the edit button text should be "編集"

### Scenario 5: Dashboard handles missing expense ID gracefully
**Status**: [ ] Not started

Given I open the dashboard with URL "?edit=nonexistent_id"
When the dashboard loads
Then it should not crash
And it should show a message "Expense not found"
And I should still be able to view my expense list

### Scenario 6: Dashboard auto-scrolls to expense when edit parameter present
**Status**: [ ] Not started

Given I have 100 expenses in my list
And expense "Lunch $85" with ID "exp_123" is at position 50
When I open dashboard with "?edit=exp_123"
Then the page should scroll to expense "Lunch $85"
And the expense should be opened in edit mode

## Domain Boundaries

### Affected Components:
1. **LINE Messenger Adapter** (Infrastructure Layer)
   - `internal/adapter/messenger/line/flex/expense.go` - Generate flex message with edit button
   - `internal/adapter/messenger/line/handler.go` - Pass expense ID to flex builder

2. **Frontend Dashboard** (Presentation Layer)
   - `frontend/dashboard/src/app/[locale]/expenses/page.tsx` - Handle query parameters
   - `frontend/dashboard/src/components/ExpenseList.tsx` - Auto-open edit mode

3. **I18n Support** (Infrastructure Layer)
   - `internal/i18n/locales/*.json` - Add translation keys for edit button

## Acceptance Criteria
- [ ] All scenarios pass with tests
- [ ] Backend tests pass: `go test ./internal/adapter/messenger/line/flex/...`
- [ ] Frontend tests pass: `npm test` (if applicable)
- [ ] Manual E2E test: Record expense → Click edit → Dashboard opens with expense in edit mode
- [ ] Works in all supported languages (zh-TW, en, ja, es)
- [ ] No regression in existing expense recording flow
