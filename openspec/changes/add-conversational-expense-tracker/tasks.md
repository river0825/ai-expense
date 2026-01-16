# Implementation Tasks

## Phase 0: Project Setup
- [ ] 0.1 Initialize Go module and set up project structure (cmd/, internal/, migrations/)
- [ ] 0.2 Add dependencies: go-sqlite3, chi (or net/http), line-bot-sdk, google-generative-ai
- [ ] 0.3 Create SQLite database and run migrations (users, expenses, categories, category_keywords, metrics tables)
- [ ] 0.4 Set up configuration management (env variables for LINE token, Gemini API key, database path)

## Phase 1: Domain Layer (Entities & Interfaces)
- [ ] 1.1 Define domain models (Expense, Category, User structs)
- [ ] 1.2 Define repository interfaces (ExpenseRepository, CategoryRepository, UserRepository)
- [ ] 1.3 Define service interfaces (MessengerService, ConversationParser, AIService, MetricsService)
- [ ] 1.4 Set up dependency injection container (wire or manual)

## Phase 2: Repository Layer (SQLite Implementation)
- [ ] 2.1 Implement UserRepository (SQLite): Create, Read, check existence
- [ ] 2.2 Implement ExpenseRepository (SQLite): Create, Read, Update, Delete, Query by date/category
- [ ] 2.3 Implement CategoryRepository (SQLite): CRUD for categories and keyword mappings
- [ ] 2.4 Implement MetricsRepository (SQLite): Query for DAU, expense summaries, category trends
- [ ] 2.5 Implement migrations runner (run .sql files on startup)
- [ ] 2.6 Add indexes and optimize queries for common patterns

## Phase 2.5: AI Service Layer
- [ ] 2.5.1 Define AIService interface (ParseExpense, SuggestCategory methods)
- [ ] 2.5.2 Implement GeminiAI (using google-generative-ai SDK)
  - [ ] 2.5.2a Implement ParseExpense method (prompt engineering for conversation parsing)
  - [ ] 2.5.2b Implement SuggestCategory method (category inference)
  - [ ] 2.5.2c Add caching layer for parsed results
  - [ ] 2.5.2d Add fallback to regex parsing on errors/timeouts
- [ ] 2.5.3 Scaffold future implementations (ClaudeAI, OpenAI interface stubs)
- [ ] 2.5.4 Set up configuration to select AI provider

## Phase 3: Use Cases (Business Logic)
- [ ] 3.1 Implement AutoSignupUseCase
  - [ ] 3.1a Check if user exists
  - [ ] 3.1b Create user record with messenger_type
  - [ ] 3.1c Initialize default categories for new user
  - [ ] 3.1d Handle race conditions (idempotent)
- [ ] 3.2 Implement ParseConversationUseCase (extract expenses from text using AI)
  - [ ] 3.2a Call AIService.ParseExpense()
  - [ ] 3.2b Relative date parsing (昨天, 上週, 上個月, etc.)
  - [ ] 3.2c Batch parsing (multiple items in one message)
  - [ ] 3.2d Validation (ensure amount + description exist)
- [ ] 3.3 Implement CreateExpenseUseCase (with AI-powered category suggestion)
  - [ ] 3.3a Parse input using ParseConversationUseCase
  - [ ] 3.3b Call AIService.SuggestCategory()
  - [ ] 3.3c Save expense to repository
- [ ] 3.4 Implement GetExpenseUseCase (list, filter by date/category)
- [ ] 3.5 Implement UpdateExpenseUseCase
- [ ] 3.6 Implement DeleteExpenseUseCase
- [ ] 3.7 Implement ManageCategoryUseCase (CRUD + AI suggestion engine)
- [ ] 3.8 Implement GenerateReportUseCase (summary + breakdown)
- [ ] 3.9 Implement MetricsAggregatorUseCase
  - [ ] 3.9a Track DAU (daily active users)
  - [ ] 3.9b Aggregate expenses (daily/weekly/monthly totals)
  - [ ] 3.9c Category trend analysis
  - [ ] 3.9d User growth metrics

## Phase 4: HTTP Adapter Layer (REST API)
- [ ] 4.1 Implement REST server with routing (chi or net/http mux)
- [ ] 4.2 Implement user handlers (POST /api/users/auto-signup)
- [ ] 4.3 Implement expense handlers (POST/GET/PUT/DELETE /api/expenses)
- [ ] 4.4 Implement parse handler (POST /api/expenses/parse)
- [ ] 4.5 Implement category handlers (GET/POST /api/categories)
- [ ] 4.6 Implement report handlers (GET /api/reports/summary, /breakdown)
- [ ] 4.7 Implement metrics handlers (GET /api/metrics/dau, /expenses-summary, /category-trends, /growth)
- [ ] 4.8 Add authentication for metrics endpoints
- [ ] 4.9 Add request validation and error response formatting
- [ ] 4.10 Add logging and request tracing

## Phase 5: Messenger Adapter Layer (LINE)
- [ ] 5.1 Implement LINE webhook handler (receive messages)
- [ ] 5.2 Implement LINE signature verification (HMAC-SHA256)
- [ ] 5.3 Implement auto-signup flow in LINE adapter (call AutoSignupUseCase)
- [ ] 5.4 Implement LINE formatter (convert API responses to LINE messages)
- [ ] 5.5 Implement LINE client (send messages via API)
- [ ] 5.6 Wire LINE adapter to core APIs (auto-signup → parse → create expense → format → send)
- [ ] 5.7 Handle LINE-specific constraints (2500 char limit, message batching)
- [ ] 5.8 Scaffold Telegram adapter (implement same interfaces for future use)

## Phase 6: Testing & Quality
- [ ] 6.1 Unit tests for use cases (mock repositories, AIService)
- [ ] 6.2 Unit tests for AI integration (mock Gemini responses)
- [ ] 6.3 Unit tests for parser (conversation → expense structs)
- [ ] 6.4 Integration tests for repository layer (SQLite)
- [ ] 6.5 Integration tests for HTTP handlers (mock use cases)
- [ ] 6.6 Integration tests for metrics aggregation
- [ ] 6.7 End-to-end tests with LINE webhook (optional: sandbox)
- [ ] 6.8 Error handling and edge case validation
- [ ] 6.9 Cost testing (monitor Gemini API usage, validate caching)

## Phase 7: Metrics & Monitoring
- [ ] 7.1 Implement metrics dashboard data collection
- [ ] 7.2 Add monitoring for Gemini API costs and latency
- [ ] 7.3 Set up alerts for API failures

## Phase 8: Future Messenger Support (Telegram)
- [ ] 8.1 Create Telegram adapter skeleton (TelegramMessenger interface impl)
- [ ] 8.2 Add Telegram handler (receive messages)
- [ ] 8.3 Implement Telegram formatter and client

## Phase 9: Deployment & Documentation
- [ ] 9.1 Build single binary and test locally
- [ ] 9.2 Create deployment instructions (Docker optional)
- [ ] 9.3 Set up LINE webhook URL and test integration
- [ ] 9.4 Write API documentation (OpenAPI/Swagger optional)
- [ ] 9.5 Write operator guide (how to switch AI providers)
- [ ] 9.6 Write user documentation
