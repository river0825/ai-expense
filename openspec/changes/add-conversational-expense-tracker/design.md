# Design: Conversational Expense Tracking System

## Context
Building an expense tracking system that operates through natural language conversation, starting with LINE. The system needs to:
- Parse human-readable expense input (descriptions + amounts)
- Automatically extract and parse dates from context
- Intelligently categorize expenses with user override capability
- Generate readable reports on demand
- Maintain conversation state per user
- Provide REST API as primary interface
- Support multiple messengers (LINE first, Telegram and others later)
- Use clean architecture for maintainability and extensibility

## Goals
- **Primary**: Enable frictionless expense entry via conversation (reduce effort vs. traditional UIs)
- **Secondary**: Smart categorization (suggest, but allow override or custom categories)
- **Non-Goal**: Mobile app, advanced analytics, receipt OCR, multi-user families (Phase 1)

## Architecture Overview

### Tech Stack
- **Language**: Go (for performance, concurrency, and simplicity)
- **Database**: SQLite (simple, zero-config, embedded)
- **HTTP Framework**: Go standard library or minimal framework (net/http)
- **Architecture Pattern**: Clean Architecture (Entities → Use Cases → Interface Adapters → Frameworks)

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────┐
│  Frameworks & Drivers                           │
│  ├── HTTP Server (REST API)                     │
│  ├── LINE Bot Adapter                           │
│  ├── Telegram Bot Adapter (future)              │
│  └── Database (SQLite)                          │
├─────────────────────────────────────────────────┤
│  Interface Adapters (Controllers, Gateways)     │
│  ├── ExpenseController                          │
│  ├── CategoryController                         │
│  ├── ReportController                           │
│  ├── MessengerGateway (LINE, Telegram)          │
│  └── RepositoryImpl (SQLite)                     │
├─────────────────────────────────────────────────┤
│  Use Cases (Business Logic)                     │
│  ├── CreateExpenseUseCase                       │
│  ├── GetExpenseUseCase                          │
│  ├── GenerateReportUseCase                      │
│  ├── ManageCategoryUseCase                      │
│  └── ParseConversationUseCase                   │
├─────────────────────────────────────────────────┤
│  Entities (Domain Models, Interfaces)           │
│  ├── Expense                                    │
│  ├── Category                                   │
│  ├── User                                       │
│  ├── ExpenseRepository (interface)              │
│  ├── CategoryRepository (interface)             │
│  └── MessengerService (interface)               │
└─────────────────────────────────────────────────┘
```

### API-First Design

The system exposes REST API as the primary interface. Messengers (LINE, Telegram, etc.) act as clients calling the API.

**Key API Endpoints**:
- `POST /api/expenses` - Create expense
- `GET /api/expenses` - List expenses (with filters)
- `PUT /api/expenses/{id}` - Update expense
- `DELETE /api/expenses/{id}` - Delete expense
- `POST /api/expenses/parse` - Parse conversation text
- `GET /api/categories` - List categories
- `POST /api/categories` - Create custom category
- `GET /api/reports/summary` - Get summary report
- `GET /api/reports/breakdown` - Get category breakdown

**Authentication**: User ID from messenger (LINE user ID, Telegram user ID, etc.)

### System Components (Use Cases)

1. **Auto Signup** (user-signup)
   - Automatic user registration when first message received
   - Create user record with messenger type and user ID
   - Initialize default categories for new user
   - `POST /api/users/auto-signup` endpoint (called by messenger adapter)

2. **AI Service** (ai-service)
   - Pluggable AI interface for conversation parsing and categorization
   - Gemini 2.5 lite implementation (initial)
   - Replaceable with other models (Claude, OpenAI, etc.) without changing use cases
   - Cached responses to avoid duplicate API calls
   - `internal/ai/` package with interface-based design

3. **Conversation Parser** (conversation-parsing)
   - AI-powered parser to extract expenses from natural language
   - Uses Gemini 2.5 lite to understand context and extract amounts/descriptions
   - Date parser for relative dates (昨天, 上週, etc.)
   - Validation that each expense has amount + description
   - Fallback to regex parsing if AI unavailable
   - `POST /api/expenses/parse` endpoint

4. **Expense Management** (expense-management)
   - CRUD operations for expense records via REST API
   - Query by date range, category, or all
   - Stores: date, amount, description, category, user_id
   - Repository pattern for SQLite abstraction

5. **Category Manager** (category-management)
   - Maintains list of categories (default + custom)
   - Uses Gemini to intelligently suggest categories
   - Allows CRUD on custom categories via REST API
   - Tracks category usage for deletion constraints

6. **Report Generator** (reporting)
   - Generates summaries (total by date range)
   - Generates breakdowns (total by category)
   - Returns JSON from API; formatters handle messenger-specific rendering
   - Respects messenger character limits (LINE's 2500 chars, etc.)

7. **Metrics Aggregator** (dashboard-metrics)
   - Tracks Daily Active Users (DAU)
   - Tracks total expenses per day/week/month
   - Aggregates by category for trend analysis
   - Calculates growth metrics (new users, expense growth)
   - `GET /api/metrics/dau` - Daily active users
   - `GET /api/metrics/expenses-summary` - Expense totals
   - `GET /api/metrics/category-trends` - Category breakdown
   - `GET /api/metrics/growth` - Growth metrics

8. **Messenger Adapters** (line-integration, future telegram-integration)
   - Auto-register user on first message
   - Receive messages from LINE Messaging API (webhook)
   - Parse conversation via `/api/expenses/parse`
   - Call `/api/expenses` to create/update/delete
   - Call `/api/reports/*` to generate reports
   - Format API responses for messenger platform
   - Send responses back to user via each messenger's API

### Messenger Adapter Pattern

```
┌─────────────────────────────────────┐
│  LINE User Chat                     │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│  LINE Webhook Handler               │
│  ├── Verify signature               │
│  ├── Parse LINE event               │
│  └── Call Messenger Gateway         │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│  MessengerGateway (Interface)       │
│  ├── HandleMessage(ctx, userID,    │
│  │    messageText)                 │
│  └── SendMessage(ctx, userID,      │
│       messageText)                 │
└──────────┬──────────────────────────┘
           │
      ┌────┴────┐
      ▼         ▼
┌──────────┐  ┌──────────────────────┐
│REST API  │  │LINE Messenger Impl   │
│Endpoints │  │(sends via LINE API)  │
└──────────┘  └──────────────────────┘
```

**Adapter Pattern Benefits**:
- New messengers (Telegram, Discord, Slack) add new adapter implementations
- Core business logic unchanged
- Each messenger's specific features (buttons, rich messages) handled by adapter
- Easy to test: mock MessengerGateway for unit tests

### AI Service Plugin Architecture

```
┌─────────────────────────────────────┐
│  Use Cases                          │
│  ├── ParseConversationUseCase       │
│  └── ManageCategoryUseCase          │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│  AIService (Interface)              │
│  ├── ParseExpense()                 │
│  └── SuggestCategory()              │
└──────────┬──────────────────────────┘
           │
      ┌────┴────────────────┐
      ▼                     ▼
┌──────────────┐  ┌──────────────────────┐
│ GeminiAI     │  │ Future: ClaudeAI,    │
│ (2.5 lite)   │  │ OpenAI, Local LLM    │
└──────────────┘  └──────────────────────┘
```

**AI Abstraction Benefits**:
- Swap Gemini for Claude, OpenAI, or local LLM without changing use cases
- Easy to mock for testing
- Cost optimization: can switch providers or use cheaper models later
- Fallback strategy: if AI unavailable, fall back to regex parsing

**Go Implementation**:
```go
// Domain interface (independent of implementation)
type AIService interface {
  ParseExpense(ctx context.Context, text string) (*ParsedExpense, error)
  SuggestCategory(ctx context.Context, description string) (string, error)
}

// Gemini implementation
type GeminiAI struct {
  client *genai.Client
  model  string // "gemini-2.5-lite"
}

// Swappable implementations
type ClaudeAI struct { /* ... */ }
type LocalLLM struct { /* ... */ }
```

### Auto Signup Flow

```
User sends message in LINE
      ↓
LINE Webhook Handler receives event
      ↓
Check if user exists in database
      ├─ Yes → Proceed to handle message
      ├─ No → Call AutoSignupUseCase
      │   ├── Create user record (user_id, messenger_type, created_at)
      │   └── Initialize default categories for user
      └─ Return (user now exists)
      ↓
Handle message (parse, create expense, etc.)
```

**Benefits**:
- Zero friction: users don't need to sign up separately
- Automatic user tracking for metrics
- Supports multi-messenger: same flow for Telegram, Discord, etc.

### Data Model (SQLite Schema)

```sql
-- User profiles
CREATE TABLE users (
  user_id TEXT PRIMARY KEY,        -- LINE/Telegram/etc user ID
  messenger_type TEXT NOT NULL,    -- 'line', 'telegram', etc.
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Expense records
CREATE TABLE expenses (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  description TEXT NOT NULL,
  amount DECIMAL NOT NULL,
  category_id TEXT,
  expense_date DATE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id),
  FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- Categories
CREATE TABLE categories (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  is_default BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id),
  UNIQUE(user_id, name)
);

-- Category suggestion keywords
CREATE TABLE category_keywords (
  id TEXT PRIMARY KEY,
  category_id TEXT NOT NULL,
  keyword TEXT NOT NULL,
  priority INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (category_id) REFERENCES categories(id),
  UNIQUE(category_id, keyword)
);

-- Indexes for performance
CREATE INDEX idx_expenses_user_date ON expenses(user_id, expense_date);
CREATE INDEX idx_expenses_user_category ON expenses(user_id, category_id);
CREATE INDEX idx_categories_user ON categories(user_id);
```

**Design Rationale**:
- TEXT for IDs (use UUID or hashid in Go)
- DECIMAL for amounts (avoid floating point issues)
- DATE for expense_date (easier filtering by date range)
- Separate keyword table for flexible category matching
- Indexes on common query patterns (per-user, by date, by category)

### Decision: Language & Framework (Go + SQLite)
**Decision**: Use Go for the backend, SQLite for persistence

**Rationale**:
- **Go**: Fast compilation, simple concurrency, minimal dependencies, excellent standard library
- **SQLite**: Zero-config database, perfect for single-user/small-scale apps, no server setup
- **Benefits**: Single binary deployment, easy local development, no external DB to manage
- **Trade-offs**: None for Phase 1; can migrate to PostgreSQL later if needed

**Go Specifics**:
- Use `database/sql` with `github.com/mattn/go-sqlite3` driver
- Minimal HTTP framework (standard `net/http` or `chi` for routing)
- Dependency injection for clean architecture

### Decision: Clean Architecture Pattern
**Decision**: Implement Clean Architecture (onion layers) in Go

**Rationale**:
- **Testability**: Mock repositories, use cases without database
- **Maintainability**: Clear separation of concerns
- **Extensibility**: Easy to add new messengers (Telegram, Discord) as adapters
- **Independence**: Business logic independent of frameworks

**Go Project Structure**:
```
aiexpense/
├── cmd/                           # Application entry points
│   └── server/main.go
├── internal/
│   ├── domain/                    # Entities & interfaces
│   │   ├── models.go              # Expense, Category, User
│   │   ├── repositories.go        # Repository interfaces
│   │   └── services.go            # Service interfaces
│   ├── usecase/                   # Business logic
│   │   ├── create_expense.go
│   │   ├── get_expenses.go
│   │   ├── generate_report.go
│   │   ├── manage_category.go
│   │   └── parse_conversation.go
│   ├── adapter/                   # Interface adapters
│   │   ├── http/                  # REST API handlers
│   │   │   ├── expense_handler.go
│   │   │   ├── category_handler.go
│   │   │   ├── report_handler.go
│   │   │   └── messenger_handler.go
│   │   ├── repository/            # Database implementations
│   │   │   ├── sqlite/
│   │   │   │   ├── expense_repo.go
│   │   │   │   └── category_repo.go
│   │   │   └── sqlite.go
│   │   └── messenger/             # Messenger adapters
│   │       ├── line/
│   │       │   ├── handler.go
│   │       │   ├── formatter.go
│   │       │   └── client.go
│   │       └── telegram/          # Future
│   └── config/                    # Configuration
│       └── config.go
├── migrations/                    # SQL migrations
│   └── 001_init_schema.up.sql
├── tests/                         # Integration tests
│   └── integration_test.go
└── go.mod / go.sum
```

### Decision: API-First Design
**Decision**: Build REST API as primary interface; messengers are API clients

**Rationale**:
- **Flexibility**: Future web UI, mobile app, or direct API users all use same endpoints
- **Testability**: Easy to test endpoints with curl/Postman before integrating messengers
- **Scalability**: Messengers can scale independently from API
- **Separation**: Business logic doesn't depend on messenger specifics

**API Contracts** (JSON):
```go
// POST /api/expenses/parse
{
  "user_id": "U1234567890abcdef",
  "text": "早餐$20午餐$30加油$200"
}

Response:
{
  "expenses": [
    {
      "description": "早餐",
      "amount": 20,
      "suggested_category": "Food",
      "date": "2024-01-16"
    },
    ...
  ],
  "status": "success"
}

// POST /api/expenses
{
  "user_id": "U1234567890abcdef",
  "description": "早餐",
  "amount": 20,
  "category_id": "cat_001",
  "date": "2024-01-16"
}

Response:
{
  "id": "exp_001",
  "status": "created",
  "message": "早餐 20元，已儲存"
}
```

### Decision: Messenger Adapter Pattern
**Decision**: Use adapter pattern to support multiple messengers

**Rationale**:
- **Extensibility**: Add Telegram, Discord, Slack without changing core code
- **Isolation**: Each messenger's quirks (rate limits, message limits, auth) isolated
- **Testing**: Mock messenger implementation for testing core logic

**Implementation**:
```go
// Domain interface
type MessengerService interface {
  SendMessage(ctx context.Context, userID, message string) error
  HandleWebhook(ctx context.Context, body []byte) error
}

// LINE adapter
type LineMessenger struct {
  channelToken string
  apiClient    *line.Client
}

// Telegram adapter (future)
type TelegramMessenger struct {
  botToken string
  apiClient *telegramapi.Client
}
```

### Decision: Single Message vs Multiple Messages
**Decision**: Deliver all responses in single consolidated messages

**Rationale**:
- Reduces noise in LINE chat
- Improves UX: user sees complete result without scrolling
- LINE's API charges per message sent; consolidation reduces costs
- Easier to correlate request → response

**Trade-off**: Very large reports may need pagination (handled in future iteration)

### Decision: Category Suggestion via Keywords
**Decision**: Use keyword matching with default categories, allow custom categories

**Rationale**:
- Simple, deterministic (no ML dependency)
- Handles common cases (早餐 → Food, 加油 → Transport)
- Users can override and create custom categories on-the-fly
- Keyword list easily maintained in database

**Alternatives considered**:
- Machine learning classification: Too complex for Phase 1, overkill for expense categorization
- Manual categorization only: High friction, defeats purpose of "conversational"
- LLM-based suggestion: Potential latency, external dependency, cost

### Decision: Relative Date Parsing
**Decision**: Support relative dates (昨天, 上週, 上個月) and default to today

**Rationale**:
- Matches natural user speech patterns
- Reduces need for date picker UIs
- Covers 90% of use cases (most expenses logged today or yesterday)

**Implementation**: Simple string matching for common patterns, graceful fallback to today

### Decision: Single User Isolation
**Decision**: Isolate all expense data by LINE user_id

**Rationale**:
- Aligns with LINE's per-chat context model
- Simplifies initial implementation (no user management)
- Easily scalable to family/multi-user sharing later

### Decision: Database Schema Simplicity
**Decision**: SQLite/PostgreSQL with simple schema; no complex transactions

**Rationale**:
- Expense tracking doesn't require distributed transactions
- Simple schema easier to reason about and migrate
- Avoids premature complexity

## Open Questions

1. **Keyword Matching Accuracy**: Should we support fuzzy matching for category suggestions, or exact keywords only? (Current: exact keywords)
2. **Timezone Handling**: How should "昨天" be interpreted across timezones? (Current: user's local timezone)
3. **Currency Support**: Single currency (TWD) or multi-currency? (Current: Phase 1 = single currency)
4. **Expense Limits**: Any validation on maximum/minimum expense amounts? (Current: no limits)

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Parser misses valid expense format | User frustration | Broad regex patterns; ask for clarification if unsure |
| LINE webhook timeout (>3s) | Message lost | Async job queue for processing; acknowledge immediately |
| User privacy (expense data on server) | Data breach | Encrypt at rest, HTTPS-only, delete policies |
| Rate limiting (users send many messages quickly) | Dropped messages | Queue incoming messages, process sequentially per user |
| Keyword collision (keyword matches multiple categories) | Wrong categorization | Use longest-match rule; allow user override |

## Migration & Rollout

1. **Deployment**: Deploy bot to staging, validate webhook connectivity
2. **Soft launch**: Roll out to small user group; monitor error logs
3. **Scale**: Expand to all users after 1 week of stability
4. **Feedback**: Collect user feedback on category suggestions and keywords

## Testing Strategy

- **Unit tests**: Parser logic, category matching, date parsing
- **Integration tests**: Expense CRUD, category CRUD, report generation
- **E2E tests**: LINE webhook → parser → save → report (using test user)
- **Load testing**: Concurrent users, message queue throughput

## Future Enhancements

- Budget alerts/limits
- Recurring expenses
- Expense editing/deletion via conversation
- Receipt image upload & OCR
- Multi-user family accounts
- Analytics (pie charts, spending trends)
- Data export (CSV, PDF)
