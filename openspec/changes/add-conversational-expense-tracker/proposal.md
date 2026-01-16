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
