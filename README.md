# AIExpense - Conversational Expense Tracking System

A frictionless expense tracking bot that operates through natural language conversation. Users chat with the bot on LINE (with support for Telegram and other messengers in the future) to log expenses and generate reports.

## Features

- **Auto-Signup**: Users automatically register when they first message the bot
- **AI-Powered Parsing**: Uses Gemini 2.5 lite to understand natural language expense descriptions
- **Smart Categorization**: AI suggests expense categories based on descriptions
- **REST API-First**: Core logic exposed via REST API, messengers act as clients
- **Multi-Messenger Support**: Start with LINE, easily extend to Telegram, Discord, etc.
- **Business Metrics Dashboard**: Track DAU, expenses, category trends, and user growth
- **Pluggable AI**: Swap Gemini for Claude, OpenAI, or local LLM via configuration

## Architecture

The system follows **Clean Architecture** with four layers:

```
Frameworks & Drivers (HTTP, LINE, SQLite)
    ↓
Interface Adapters (REST handlers, Repository impl, Messenger adapters)
    ↓
Use Cases (Business logic)
    ↓
Entities (Domain models, Interfaces)
```

### Project Structure

```
aiexpense/
├── cmd/server/             # Application entry point
├── internal/
│   ├── domain/             # Models and interfaces
│   ├── usecase/            # Business logic
│   ├── adapter/
│   │   ├── http/           # REST API handlers
│   │   ├── repository/     # SQLite implementations
│   │   └── messenger/      # LINE, Telegram adapters
│   ├── ai/                 # AI service abstraction (Gemini)
│   └── config/             # Configuration
├── migrations/             # Database schema
└── tests/                  # Tests
```

## Tech Stack

- **Language**: Go 1.21+
- **Database**: SQLite (embedded, zero-config)
- **AI**: Google Gemini 2.5 lite (pluggable, with fallback to regex)
- **HTTP**: Go standard library (net/http)
- **Messengers**: LINE Messaging API (Telegram support ready)

## Getting Started

### Prerequisites

- Go 1.21+
- Environment variables configured:
  - `ENABLED_MESSENGERS` - Comma-separated list of enabled messengers (default: `terminal`). Options: `terminal`, `line`, `telegram`, `discord`, `slack`, `teams`, `whatsapp`.
  - `LINE_CHANNEL_TOKEN` - LINE Messaging API channel token (required if line is enabled)
  - `LINE_CHANNEL_ID` - LINE channel ID (required if line is enabled)
  - `GEMINI_API_KEY` - Google Gemini API key
  - `ADMIN_API_KEY` - (Optional) API key for metrics endpoints
  - `SERVER_PORT` - (Optional, default: 8080)
  - `DATABASE_PATH` - (Optional, default: ./aiexpense.db)

### Installation

```bash
git clone <repo>
cd aiexpense
go build ./cmd/server
```

### Running

#### Quick Start (Terminal Mode)

```bash
# Runs with terminal messenger by default
# No external credentials required
./server
```

#### Production Mode (LINE)

```bash
export ENABLED_MESSENGERS=line
export LINE_CHANNEL_TOKEN=<your_token>
export LINE_CHANNEL_ID=<your_id>
export GEMINI_API_KEY=<your_api_key>
./server
```

The server will start on `http://localhost:8080`.

#### Using Terminal Messenger

If running in terminal mode (`ENABLED_MESSENGERS=terminal`), you can interact with the bot via API:

```bash
# Send a message
curl -X POST http://localhost:8080/api/chat/terminal \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test_user", "message": "早餐$20"}'

# Check user stats
curl "http://localhost:8080/api/chat/terminal/user?user_id=test_user"
```

## API Endpoints

### User Management
- `POST /api/users/auto-signup` - Auto-register user

### Expense Management
- `POST /api/expenses/parse` - Parse conversation text
- `POST /api/expenses` - Create expense
- `GET /api/expenses` - List user's expenses

### Category Management
- `GET /api/categories` - List categories

### Metrics (requires `X-API-Key` header)
- `GET /api/metrics/dau` - Daily active users
- `GET /api/metrics/expenses-summary` - Expense totals by date
- `GET /api/metrics/growth` - User growth metrics

### Webhooks
- `POST /webhook/line` - LINE Messaging API webhook

## Database Schema

### Users Table
- `user_id` (TEXT, PRIMARY KEY) - Messenger user ID
- `messenger_type` (TEXT) - 'line', 'telegram', etc.
- `created_at` (TIMESTAMP)

### Expenses Table
- `id` (TEXT, PRIMARY KEY)
- `user_id` (TEXT, FK)
- `description` (TEXT)
- `amount` (DECIMAL)
- `category_id` (TEXT, FK)
- `expense_date` (DATE)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### Categories Table
- `id` (TEXT, PRIMARY KEY)
- `user_id` (TEXT, FK)
- `name` (TEXT)
- `is_default` (BOOLEAN)
- `created_at` (TIMESTAMP)

### Default Categories
- Food
- Transport
- Shopping
- Entertainment
- Other

## Usage Examples

### Creating an expense via REST API

```bash
curl -X POST http://localhost:8080/api/expenses \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "U1234567890abcdef",
    "description": "早餐",
    "amount": 20,
    "date": "2024-01-16T00:00:00Z"
  }'
```

### Parsing conversation text

```bash
curl -X POST http://localhost:8080/api/expenses/parse \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "U1234567890abcdef",
    "text": "早餐$20午餐$30加油$200"
  }'
```

### Getting metrics

```bash
curl http://localhost:8080/api/metrics/dau \
  -H "X-API-Key: <admin_api_key>"
```

### Via LINE Chat

Simply send a message like:
```
早餐$20午餐$30加油$200
```

Bot will respond with:
```
早餐 20元 [Food]，已儲存
午餐 30元 [Food]，已儲存
加油 200元 [Transport]，已儲存
```

## AI Service Configuration

### Switch AI Providers

Set `AI_PROVIDER` environment variable:

```bash
# Using Gemini (default)
export AI_PROVIDER=gemini
export GEMINI_API_KEY=<key>

# Using Claude (future)
export AI_PROVIDER=claude
export CLAUDE_API_KEY=<key>

# Using OpenAI (future)
export AI_PROVIDER=openai
export OPENAI_API_KEY=<key>
```

The system automatically swaps the AI implementation without code changes.

## Fallback Behavior

If AI service is unavailable, the system falls back to regex-based parsing:
- Pattern: `description$amount` (e.g., "早餐$20")
- Still extracts amounts and descriptions correctly
- User experience unaffected

## Next Steps (Future Phases)

### Phase 6: Testing & Quality
- Unit tests for all use cases
- Integration tests for repository layer
- E2E tests with LINE webhook
- Cost monitoring for AI API

### Phase 7: Metrics & Monitoring
- Dashboard for business metrics
- Cost tracking and alerts
- User activity insights

### Phase 8: Telegram Integration
- Telegram bot adapter
- Multi-messenger message sync (optional)

### Phase 9: Advanced Features
- Expense editing/deletion via conversation
- Receipt image OCR
- Budget alerts
- Multi-user family sharing
- Data export (CSV, PDF)

## Development

### Running Tests

```bash
go test ./...
```

### Building Release Binary

```bash
go build -o aiexpense ./cmd/server
```

## License

MIT

## Contributing

Contributions welcome! Please follow clean architecture principles and add tests for new features.

## Support

For issues or feature requests, create an GitHub issue.
