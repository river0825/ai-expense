# Implementation Tasks (TDD/BDD/DDD)

## Phase 0: Project Setup
- [x] 0.1 Initialize Go module and set up project structure (cmd/, internal/, migrations/)
- [x] 0.2 Add dependencies: go-sqlite3, chi (or net/http), line-bot-sdk, google-generative-ai
- [x] 0.3 Create SQLite database and run migrations (users, expenses, categories, category_keywords, metrics tables)
- [x] 0.4 Set up configuration management (env variables for LINE token, Gemini API key, database path)

## Phase 1: Domain Layer (DDD - Entities & Interfaces)
- [x] 1.1 Define domain models (Expense, Category, User structs with value objects)
- [x] 1.2 Define repository interfaces (ExpenseRepository, CategoryRepository, UserRepository)
- [x] 1.3 Define service interfaces (MessengerService, ConversationParser, AIService, MetricsService)
- [x] 1.4 Set up dependency injection container (wire or manual)

## Phase 2: Repository Layer (SQLite Implementation)
- [x] 2.1 Implement UserRepository (SQLite): Create, Read, check existence
- [x] 2.2 Implement ExpenseRepository (SQLite): Create, Read, Update, Delete, Query by date/category
- [x] 2.3 Implement CategoryRepository (SQLite): CRUD for categories and keyword mappings
- [x] 2.4 Implement MetricsRepository (SQLite): Query for DAU, expense summaries, category trends
- [x] 2.5 Implement migrations runner (run .sql files on startup)
- [x] 2.6 Add indexes and optimize queries for common patterns

## Phase 2.5: AI Service Layer
- [x] 2.5.1 Define AIService interface (ParseExpense, SuggestCategory methods)
- [x] 2.5.2 Implement GeminiAI (using google-generative-ai SDK)
  - [x] 2.5.2a Implement ParseExpense method (prompt engineering for conversation parsing)
  - [x] 2.5.2b Implement SuggestCategory method (category inference)
  - [x] 2.5.2c Add caching layer for parsed results
  - [x] 2.5.2d Add fallback to regex parsing on errors/timeouts
- [x] 2.5.3 Scaffold future implementations (ClaudeAI, OpenAI interface stubs)
- [x] 2.5.4 Set up configuration to select AI provider

## Phase 3: Use Cases (Business Logic)
- [x] 3.1 Implement AutoSignupUseCase
  - [x] 3.1a Check if user exists
  - [x] 3.1b Create user record with messenger_type
  - [x] 3.1c Initialize default categories for new user
  - [x] 3.1d Handle race conditions (idempotent)
- [x] 3.2 Implement ParseConversationUseCase (extract expenses from text using AI)
  - [x] 3.2a Call AIService.ParseExpense()
  - [x] 3.2b Relative date parsing (昨天, 上週, 上個月, etc.)
  - [x] 3.2c Batch parsing (multiple items in one message)
  - [x] 3.2d Validation (ensure amount + description exist)
- [x] 3.3 Implement CreateExpenseUseCase (with AI-powered category suggestion)
  - [x] 3.3a Parse input using ParseConversationUseCase
  - [x] 3.3b Call AIService.SuggestCategory()
  - [x] 3.3c Save expense to repository
- [x] 3.4 Implement GetExpenseUseCase (list, filter by date/category)
- [x] 3.5 Implement UpdateExpenseUseCase
- [x] 3.6 Implement DeleteExpenseUseCase
- [x] 3.7 Implement ManageCategoryUseCase (CRUD + AI suggestion engine)
- [x] 3.8 Implement GenerateReportUseCase (summary + breakdown)
- [x] 3.9 Implement MetricsAggregatorUseCase
  - [x] 3.9a Track DAU (daily active users)
  - [x] 3.9b Aggregate expenses (daily/weekly/monthly totals)
  - [x] 3.9c Category trend analysis
  - [x] 3.9d User growth metrics

## Phase 4: HTTP Adapter Layer (REST API)
- [x] 4.1 Implement REST server with routing (chi or net/http mux)
- [x] 4.2 Implement user handlers (POST /api/users/auto-signup)
- [x] 4.3 Implement expense handlers (POST/GET/PUT/DELETE /api/expenses)
- [x] 4.4 Implement parse handler (POST /api/expenses/parse)
- [x] 4.5 Implement category handlers (GET/POST /api/categories)
- [x] 4.6 Implement report handlers (GET /api/reports/summary, /breakdown)
- [x] 4.7 Implement metrics handlers (GET /api/metrics/dau, /expenses-summary, /category-trends, /growth)
- [x] 4.8 Add authentication for metrics endpoints
- [x] 4.9 Add request validation and error response formatting
- [x] 4.10 Add logging and request tracing

## Phase 5: Messenger Adapter Layer (LINE)
- [x] 5.1 Implement LINE webhook handler (receive messages)
- [x] 5.2 Implement LINE signature verification (HMAC-SHA256)
- [x] 5.3 Implement auto-signup flow in LINE adapter (call AutoSignupUseCase)
- [x] 5.4 Implement LINE formatter (convert API responses to LINE messages)
- [x] 5.5 Implement LINE client (send messages via API)
- [x] 5.6 Wire LINE adapter to core APIs (auto-signup → parse → create expense → format → send)
- [x] 5.7 Handle LINE-specific constraints (2500 char limit, message batching)
- [x] 5.8 Scaffold Telegram adapter (implement same interfaces for future use)

## Phase 6: Testing & Quality (TDD/BDD - Test-First)
- [x] 6.1 Unit tests for use cases (mock repositories, AIService)
- [x] 6.2 Unit tests for AI integration (mock Gemini responses)
- [x] 6.3 Unit tests for parser (conversation → expense structs)
- [x] 6.4 Integration tests for repository layer (SQLite)
- [x] 6.5 Integration tests for HTTP handlers (mock use cases)
- [x] 6.6 Integration tests for metrics aggregation
- [x] 6.7 End-to-end tests with LINE webhook (optional: sandbox)
- [x] 6.8 Error handling and edge case validation
- [x] 6.9 Cost testing (monitor Gemini API usage, validate caching)

## Phase 7: Metrics & Monitoring
- [x] 7.1 Implement metrics dashboard data collection
- [x] 7.2 Add monitoring for Gemini API costs and latency
- [x] 7.3 Set up alerts for API failures

## Phase 8: Future Messenger Support (Telegram, Discord, Slack, Teams, WhatsApp)
- [x] 8.1 Create Telegram adapter skeleton (TelegramMessenger interface impl)
- [x] 8.2 Add Telegram handler (receive messages)
- [x] 8.3 Implement Telegram formatter and client
- [x] 8.4 Add Discord support (adapter, handler, client)
- [x] 8.5 Add Slack support (adapter, handler, client)
- [x] 8.6 Add Teams support (adapter, handler, client)
- [x] 8.7 Add WhatsApp support (adapter, handler, client)

## Phase 9: Deployment & Documentation
- [x] 9.1 Build single binary and test locally
- [x] 9.2 Create deployment instructions (Docker optional)
- [x] 9.3 Set up LINE webhook URL and test integration
- [x] 9.4 Write API documentation (OpenAPI/Swagger optional)
- [x] 9.5 Write operator guide (how to switch AI providers)
- [x] 9.6 Write user documentation

## Phase 16-17: Enhanced Testing & E2E Integration
- [x] Phase 16: Enhanced testing with 95%+ coverage
- [x] Phase 17: End-to-End integration tests (webhook flows, security, E2E scenarios)

## Phase 18-20: Performance & Monitoring
- [x] Phase 18: Performance testing & benchmarking (16+ benchmarks, 7 load tests)
- [x] Phase 19: Performance optimization (database, caching, async processing)
- [x] Phase 20: Production monitoring & health checks
