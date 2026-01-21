# Change: Add Conversational Expense Tracking System

## Why
Users want a frictionless way to track expenses through natural conversation, without needing to remember complex commands or navigate UI menus. A conversational, LINE-integrated bot reduces entry friction and makes expense tracking feel like chatting with a friend.

## What Changes
- **Auto-signup** - Users automatically register when first messaging the bot (no sign-up form needed)
- **AI-powered conversation parsing** - Use Gemini 2.5 lite for intelligent expense extraction from natural language
- **AI-powered category suggestions** - Gemini suggests categories based on expense descriptions (replaceable with other AI models)
- **New expense management** - Store and retrieve expense records with dates, amounts, and categories
- **New category system** - Pre-defined categories with smart defaults and user ability to add custom ones
- **New reporting** - Generate expense summaries/reports via conversation
- **Business metrics dashboard** - REST API endpoints to track DAU, total expenses, category trends, user growth
- **New LINE integration** - Connect bot to LINE Messaging API with auto-signup
- **Messenger abstraction** - Support LINE initially with easy addition of Telegram, Discord, Slack later
- **AI abstraction** - Pluggable AI service; swap Gemini for other models without code changes
- All responses delivered as **single consolidated messages** for clarity

## Tech Stack
- **Backend Language:** Go (for performance and simplicity)
- **Database:** SQLite (zero-config, embedded, perfect for single-instance apps)
- **Architecture:** Clean Architecture (domain → use cases → adapters → frameworks)
- **API Model:** REST API-first with messenger adapters (LINE, future Telegram support)
- **Framework:** Go standard library (net/http) with chi for routing (optional)

## Impact
- **Affected specs:** user-signup, conversation-parsing, expense-management, category-management, reporting, line-integration, dashboard-metrics, ai-service
- **New capabilities:** User auto-signup, AI-powered parsing, Business metrics dashboard
- **Project structure:**
  - `cmd/server/main.go` - Application entry point
  - `internal/domain/` - Entities and interfaces
  - `internal/usecase/` - Business logic (includes auto-signup, metrics aggregation)
  - `internal/adapter/` - HTTP handlers, repository implementations, messenger adapters
  - `internal/ai/` - AI service abstraction (Gemini implementation + interface for swapping)
  - `migrations/` - SQLite schema (includes user tracking for metrics)
- **Implementation phases:** 9 phases (setup → domain → repo+AI → use cases → API → LINE+metrics → testing → Telegram → deployment)

## Out of Scope (Phase 1)
- Mobile app (LINE web only)
- Advanced analytics (pie charts, trends)
- Budget alerts/limits
- Multi-user/family sharing
- Recurring expenses
- Receipt image upload

## Design Highlights
- **API-First:** Core REST API is independent of messengers; easy to add web UI or mobile app later
- **Messenger Adapters:** Clean Architecture enables adding Telegram, Discord, Slack without modifying core code
- **Single Binary:** Go compilation produces one executable; SQLite requires no external database server
- **Extensible:** Keyword-based category suggestion can be upgraded to ML later without breaking API contracts

## Gherkin Specifications (TDD/BDD)

### Feature: User Auto-Signup
```gherkin
Scenario 1: First-time user signup
[x] WHEN user sends first message to bot
[x] THEN system creates user record with messenger type
[x] AND initializes default expense categories
Tests: TestScenario_FirstTimeUserSignup ✓

Scenario 2: Existing user message
[x] WHEN existing user sends message
[x] THEN system recognizes user and processes request
[x] AND does NOT create duplicate user record
Tests: TestScenario_ExistingUserMessage ✓

Scenario 3: Multiple messenger platforms
[x] WHEN different messenger platforms connect
[x] THEN system handles each platform independently
[x] AND maintains separate user records per messenger
Tests: TestScenario_MultipleMessengerPlatforms ✓
```

### Feature: Expense Management
```gherkin
Scenario 1: Create expense from natural language
[x] WHEN user sends natural language expense description
[x] THEN system parses text to extract amount and description
[x] AND suggests appropriate category using AI
[x] AND stores expense with date, amount, category
Tests: TestScenario_CreateExpenseFromNaturalLanguage ✓

Scenario 2: List expenses by date range
[x] WHEN user requests expenses for date range
[x] THEN system returns matching expense records
[x] AND groups by category or date as requested
Tests: TestScenario_ListExpensesByDateRange ✓

Scenario 3: Update expense
[x] WHEN user modifies existing expense
[x] THEN system updates record and recalculates metrics
[x] AND maintains audit trail of changes
Tests: TestScenario_UpdateExpense ✓

Scenario 4: Delete expense
[x] WHEN user deletes own expense
[x] THEN system removes from database
[x] AND recalculates user metrics
Tests: TestScenario_DeleteExpense ✓
```

### Feature: AI-Powered Category Suggestion
```gherkin
Scenario 1: Suggest category from description
[x] WHEN AI service receives expense description
[x] THEN system suggests best matching category
[x] AND provides confidence score and alternatives
Tests: TestScenario_SuggestCategoryFromDescription ✓

Scenario 2: Learn from corrections
[x] WHEN user corrects category suggestion
[x] THEN system learns from feedback for future suggestions
[x] AND improves recommendation accuracy
Tests: TestScenario_LearnFromCorrections ✓
```

### Feature: Business Metrics Dashboard
```gherkin
Scenario 1: Daily Active Users (DAU)
[x] WHEN admin queries DAU metrics
[x] THEN system returns count of unique users per day
[x] AND shows trend over time
Tests: TestScenario_DailyActiveUsers ✓

Scenario 2: Expense Summary
[x] WHEN user requests expense summary
[x] THEN system returns total spent, by category, by time period
[x] AND provides comparison with previous periods
Tests: TestScenario_ExpenseSummary ✓

Scenario 3: Category Trends
[x] WHEN admin views category analytics
[x] THEN system shows spending by category over time
[x] AND identifies top spending categories
Tests: TestScenario_CategoryTrends ✓
```

## DDD Domain Model

### Aggregate: User
```
Entity: User
  - user_id (Value Object: UserID)
  - messenger_type (Value Object: MessengerType)
  - created_at (Value Object: Timestamp)

Repository: UserRepository
  - Create(User) -> error
  - GetByID(UserID) -> User
  - Exists(UserID) -> bool
```

### Aggregate: Expense
```
Entity: Expense
  - id (Value Object: ExpenseID)
  - user_id (Value Object: UserID)
  - amount (Value Object: Money)
  - description (Value Object: ExpenseDescription)
  - category_id (Value Object: CategoryID)
  - expense_date (Value Object: Date)
  - created_at (Value Object: Timestamp)

Domain Events:
  - ExpenseCreated(ExpenseID, UserID, Amount)
  - ExpenseUpdated(ExpenseID, Changes)
  - ExpenseDeleted(ExpenseID)

Repository: ExpenseRepository
  - Create(Expense) -> error
  - GetByID(ExpenseID) -> Expense
  - GetByUserID(UserID) -> []Expense
  - GetByUserIDAndDateRange(UserID, DateRange) -> []Expense
  - Update(Expense) -> error
  - Delete(ExpenseID) -> error
```

### Aggregate: Category
```
Entity: Category
  - id (Value Object: CategoryID)
  - user_id (Value Object: UserID)
  - name (Value Object: CategoryName)
  - is_default (bool)

Value Object: CategoryKeyword
  - keyword (string)
  - priority (int)

Repository: CategoryRepository
  - Create(Category) -> error
  - GetByID(CategoryID) -> Category
  - GetByUserID(UserID) -> []Category
  - GetByUserIDAndName(UserID, CategoryName) -> Category
  - CreateKeyword(CategoryKeyword) -> error
  - GetKeywordsByCategory(CategoryID) -> []CategoryKeyword
```

### Service: AIService (Domain Service)
```
Interface: AIService
  - ParseExpense(text, userID) -> []ParsedExpense
  - SuggestCategory(description, userID) -> CategorySuggestion

Value Objects:
  - ParsedExpense: Amount, Description, Date
  - CategorySuggestion: CategoryID, Confidence, Alternatives
```

## UseCase Design

### UseCase 1: AutoSignupUseCase
```
Input: UserID, MessengerType
Process:
  1. Check if user exists (UserRepository.Exists)
  2. If not exists:
     a. Create user record
     b. Initialize default categories
     c. Return newly created User
  3. If exists:
     a. Return existing User (idempotent)
Output: User
Errors: DatabaseError, ValidationError
```

### UseCase 2: CreateExpenseUseCase
```
Input: UserID, ExpenseText, Date (optional)
Process:
  1. Parse natural language text (AIService.ParseExpense)
  2. For each parsed expense:
     a. Suggest category (AIService.SuggestCategory)
     b. Create Expense entity
     c. Persist to repository (ExpenseRepository.Create)
     d. Emit ExpenseCreated domain event
  3. Return created Expense records
Output: []Expense
Errors: ParseError, AIServiceError, ValidationError, DatabaseError
```

### UseCase 3: GetExpensesUseCase
```
Input: UserID, DateRange (optional), CategoryID (optional)
Process:
  1. Build query filter from inputs
  2. Fetch from repository (ExpenseRepository.GetByUserIDAndDateRange)
  3. Apply category filter if specified
  4. Return sorted by date (descending)
Output: []Expense
Errors: ValidationError, DatabaseError
```

### UseCase 4: SuggestCategoryUseCase
```
Input: ExpenseDescription, UserID
Process:
  1. Call AI service (AIService.SuggestCategory)
  2. Validate suggestion against user's categories
  3. Return suggestion with confidence
Output: CategorySuggestion
Errors: AIServiceError, ValidationError
```

## Repository Interfaces (Initial In-Memory Implementation)

All repositories use In-Memory maps before database integration:

### UserRepository Interface
```go
interface UserRepository {
  Create(ctx, user *User) error
  GetByID(ctx, userID string) (*User, error)
  Exists(ctx, userID string) (bool, error)
}
```

### ExpenseRepository Interface
```go
interface ExpenseRepository {
  Create(ctx, expense *Expense) error
  GetByID(ctx, id string) (*Expense, error)
  GetByUserID(ctx, userID string) ([]*Expense, error)
  GetByUserIDAndDateRange(ctx, userID string, from, to time.Time) ([]*Expense, error)
  Update(ctx, expense *Expense) error
  Delete(ctx, id string) error
}
```

### CategoryRepository Interface
```go
interface CategoryRepository {
  Create(ctx, category *Category) error
  GetByID(ctx, id string) (*Category, error)
  GetByUserID(ctx, userID string) ([]*Category, error)
  GetByUserIDAndName(ctx, userID, name string) (*Category, error)
  CreateKeyword(ctx, keyword *CategoryKeyword) error
  GetKeywordsByCategory(ctx, categoryID string) ([]*CategoryKeyword, error)
}
```
