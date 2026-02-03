# Multi-Currency Support Design

**Date**: 2026-01-29  
**Status**: Draft  
**Estimated Effort**: 7-10 days (32 tasks)

## Overview

Enable AIExpense to support multiple currencies for expense tracking. Users can record expenses in different currencies (e.g., "早餐 300 日幣"), and the system will store both the original amount and the converted value in the user's home currency.

## Requirements Summary

| Requirement | Decision |
|-------------|----------|
| Storage Strategy | Store both original currency AND converted home currency |
| Exchange Rate Source | Daily cached rates from Frankfurter API (with interface for swapping providers) |
| Default Currency | User-configurable home currency (stored in user settings) |
| Display Format | Primary: original amount; Secondary: smaller converted value below |
| Supported Currencies | TWD, JPY, USD, EUR, CNY (5 currencies initially) |
| Ambiguous Currency | Default to user's home currency |
| i18n | Currency names use separate translations table |

## Data Model Changes

### New Tables

#### `currencies` (reference data)

| Column | Type | Description |
|--------|------|-------------|
| `code` | TEXT PK | ISO currency code: "JPY", "USD", "EUR", "TWD", "CNY" |
| `symbol` | TEXT | Currency symbol: "¥", "$", "€", "NT$" |
| `aliases` | JSONB | AI recognition patterns: `["日幣", "日元", "円", "yen"]` |
| `is_active` | BOOLEAN | Enable/disable without deleting |

#### `currency_translations` (i18n)

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL PK | |
| `currency_code` | TEXT FK | References currencies.code |
| `locale` | TEXT | "en", "zh-TW", "zh-CN", "ja" |
| `name` | TEXT | Translated name |
| **Unique** | | `(currency_code, locale)` |

#### `exchange_rates` (daily cache)

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL PK | |
| `base_currency` | TEXT | Base currency code (normalized to USD) |
| `target_currency` | TEXT | Target currency code |
| `rate` | DECIMAL | Exchange rate |
| `rate_date` | DATE | Date of the rate |
| `fetched_at` | TIMESTAMP | When rate was fetched |
| **Unique** | | `(base_currency, target_currency, rate_date)` |

### Modified Tables

#### `users` - Add columns

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `home_currency` | TEXT | "TWD" | User's preferred home currency |
| `locale` | TEXT | "zh-TW" | User's locale for i18n |

#### `expenses` - Add columns

| Column | Type | Description |
|--------|------|-------------|
| `original_amount` | DECIMAL | Amount user entered (renamed from `amount`) |
| `currency` | TEXT | Original currency code |
| `home_amount` | DECIMAL | Converted amount in user's home currency |
| `home_currency` | TEXT | User's home currency at time of entry |
| `exchange_rate` | DECIMAL | Rate used for conversion (audit trail) |

### Backfill Strategy

All existing expenses:
- `currency` = 'TWD'
- `home_currency` = 'TWD'
- `home_amount` = `original_amount`
- `exchange_rate` = 1.0

## Domain Model Changes

### New Structs

```go
// internal/domain/currency.go

type Currency struct {
    Code     string
    Symbol   string
    Aliases  []string
    IsActive bool
}

type ExchangeRate struct {
    BaseCurrency   string
    TargetCurrency string
    Rate           float64
    RateDate       time.Time
    FetchedAt      time.Time
}
```

### Modified Structs

```go
// ParsedExpense - AI response
type ParsedExpense struct {
    Description       string
    Amount            float64
    Currency          string  // Normalized ISO code: "JPY"
    CurrencyOriginal  string  // What user typed: "日幣"
    SuggestedCategory string
    Account           string
    Date              time.Time
}

// Expense - Domain model
type Expense struct {
    ID             string
    UserID         string
    Description    string
    OriginalAmount float64
    Currency       string
    HomeAmount     float64
    HomeCurrency   string
    ExchangeRate   float64
    Category       string
    Account        string
    ExpenseDate    time.Time
    CreatedAt      time.Time
}
```

### New Interfaces

```go
// CurrencyRepository
type CurrencyRepository interface {
    GetByCode(code string) (*Currency, error)
    GetAll() ([]*Currency, error)
    GetTranslation(code, locale string) (string, error)
}

// ExchangeRateRepository
type ExchangeRateRepository interface {
    GetRate(from, to string, date time.Time) (*ExchangeRate, error)
    SaveRate(rate *ExchangeRate) error
    GetLatestRates(baseCurrency string) ([]*ExchangeRate, error)
}

// ExchangeRateProvider (external API interface)
type ExchangeRateProvider interface {
    FetchRates(baseCurrency string) ([]*ExchangeRate, error)
    Provider() string
}

// ExchangeRateService
type ExchangeRateService interface {
    Convert(amount float64, from, to string, date time.Time) (converted, rate float64, err error)
    RefreshRates(ctx context.Context) error
    GetRate(from, to string, date time.Time) (float64, error)
}
```

## AI Prompt Changes

### Current Response

```json
{"description": "早餐", "amount": 300, "suggested_category": "Food"}
```

### New Response

```json
{
  "description": "早餐",
  "amount": 300,
  "currency": "JPY",
  "currency_original": "日幣",
  "suggested_category": "Food"
}
```

### Currency Resolution Priority

1. Explicit currency in message ("50 美金" → USD)
2. User's home currency (fallback when ambiguous)

## Expense Creation Flow

```
User Input: "早餐 300 日幣"
    ↓
AI Parse → {amount: 300, currency: "JPY", currency_original: "日幣"}
    ↓
Resolve Currency → JPY (from AI) or home_currency (if null)
    ↓
Fetch Exchange Rate → JPY→TWD = 0.213
    ↓
Calculate Home Amount → 300 * 0.213 = 64
    ↓
Create Expense {
    original_amount: 300,
    currency: "JPY",
    home_amount: 64,
    home_currency: "TWD",
    exchange_rate: 0.213
}
    ↓
Save & Respond: "已記錄：早餐 ¥300 (≈NT$64)"
```

## API Changes

### Modified Endpoints

| Endpoint | Changes |
|----------|---------|
| `GET /api/expenses` | Add `original_amount`, `currency`, `currency_symbol`, `home_amount`, `home_currency`, `exchange_rate` |
| `GET /api/reports/summary` | Add `by_currency` breakdown array |

### New Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /api/currencies` | List supported currencies with translations |
| `GET /api/user/settings` | Get user preferences including home_currency |
| `PUT /api/user/settings` | Update user preferences |
| `GET /api/exchange-rates` | Get current cached rates (optional, for transparency) |

## Dashboard Changes

### New Pages

| Path | Purpose |
|------|---------|
| `/[locale]/user/settings` | User settings page with home currency selector |

### Modified Components

| Component | Changes |
|-----------|---------|
| Expense List | Show original amount + smaller converted value below |
| Expense Detail Modal | Display currency info, allow editing original amount |
| Reports Page | Add currency breakdown section |

### New Components

| Component | Purpose |
|-----------|---------|
| `CurrencyAmount` | Reusable component for displaying amount with optional conversion |

```tsx
<CurrencyAmount 
  amount={300} 
  currency="JPY" 
  homeAmount={64} 
  homeCurrency="TWD" 
/>
// Renders:
// ¥300
// ≈NT$64 (smaller, gray)
```

## Exchange Rate Strategy

### Provider

- **Primary**: Frankfurter.app (free, no API key, European Central Bank data)
- **Interface**: Abstract `ExchangeRateProvider` for future provider swapping

### Refresh Schedule

- Daily at 00:05 UTC via goroutine in main.go
- Fallback: Use most recent cached rate if API unavailable

### Rate Lookup Strategy

1. Try exact date match
2. Fall back to most recent available rate (weekends/holidays)
3. Return error if no rate exists (don't silently fail)

## Risk Considerations

| Risk | Mitigation |
|------|------------|
| Frankfurter API downtime | Graceful degradation: use most recent cached rate, log warning |
| AI misinterprets currency | Keep `currency_original` for debugging, iterate on prompts |
| Exchange rate stale on weekends | Fall back to Friday's rate for Sat/Sun expenses |
| Existing expense migration issues | Run migration on staging first, verify backfill |
| User confusion about converted amounts | Clear UI labels: "≈" prefix, tooltip explaining rate source |

## Future Considerations (Phase 2)

### Trip Mode

Allow users to start a "trip" context where all expenses default to the trip's currency:

```
User: "開始日本旅行"
Bot: "已開啟日本旅行模式 🇯🇵 預設幣別：JPY"
User: "早餐 300" → Stored as 300 JPY
```

Would require:
- New `trips` table
- Trip lifecycle management (start/end)
- AI intent detection for trip commands
- Currency resolution priority update

## Files Summary

### New Files

- `migrations/0XX_add_currencies.up.sql` / `.down.sql`
- `migrations/0XX_add_user_currency.up.sql` / `.down.sql`
- `migrations/0XX_expand_expenses_currency.up.sql` / `.down.sql`
- `internal/domain/currency.go`
- `internal/adapter/repository/postgres/currency_repository.go`
- `internal/adapter/repository/postgres/exchange_rate_repository.go`
- `internal/adapter/exchangerate/frankfurter.go`
- `internal/usecase/exchange_rate_service.go`
- `dashboard/src/app/[locale]/user/settings/page.tsx`
- `dashboard/src/components/CurrencyAmount.tsx`

### Modified Files

- `internal/domain/models.go` (Expense, ParsedExpense structs)
- `internal/ai/gemini.go` (prompt and response parsing)
- `internal/usecase/expense_usecase.go` (creation flow)
- `internal/adapter/http/handler.go` (new endpoints)
- `dashboard/src/app/[locale]/expenses/page.tsx` (list display)
- `dashboard/src/app/[locale]/reports/page.tsx` (summary display)

## Step-by-Step Implementation Plan

### Phase 1 – Database & Domain Foundations

1. **Add currency infrastructure tables**
   - Create `0XX_add_currencies.{up,down}.sql` adding `currencies` and `currency_translations` with seed data for TWD/JPY/USD/EUR/CNY in `zh-TW` and `en`.
   - Add `0XX_add_exchange_rates.{up,down}.sql` (if separate) or include in same migration.
   - Run `make migrate-up` locally to validate.

2. **Add user currency preference**
   - Migration adds `home_currency TEXT DEFAULT 'TWD'` and `locale TEXT DEFAULT 'zh-TW'` to `users`.
   - Backfill existing rows via `UPDATE users SET home_currency='TWD', locale='zh-TW' WHERE home_currency IS NULL`.

3. **Expand expenses table**
   - Migration renames `amount` → `original_amount`, adds `currency`, `home_amount`, `home_currency`, `exchange_rate` (defaults `'TWD'`, same amount, `1.0`).
   - Provide down migration to reverse schema.

4. **Create domain models & repositories**
   - New `internal/domain/currency.go` defining `Currency`, `ExchangeRate`, helper methods.
   - Implement `CurrencyRepository` & `ExchangeRateRepository` in Postgres adapter with tests using existing repository patterns.

5. **Implement exchange rate provider abstraction**
   - Define `ExchangeRateProvider` interface.
   - Create `internal/adapter/exchangerate/frankfurter.go` using `net/http`, fetch daily rates (EUR base) and normalize to USD.
   - Add unit tests with httptest server fixtures.

6. **Implement ExchangeRateService**
   - New use case/service orchestrating repositories and provider (`Convert`, `GetRate`, `RefreshRates`).
   - Include caching strategy (prefer DB) and logging on provider failures.
   - Tests covering weekend fallback, cache hits, provider errors.

7. **Schedule daily refresh**
   - In `cmd/server/main.go` (or init), start goroutine with `time.NewTicker(24h)` calling `ExchangeRateService.RefreshRates`.
   - Ensure graceful shutdown via context cancellation.

### Phase 2 – AI & Use Case Integration

8. **Update AI prompt & parser**
   - Modify Gemini prompt to list supported currencies/aliases, instruct to output `currency` & `currency_original` (ISO codes uppercase).
   - Update `internal/ai/gemini.go` parsing logic, add tests for multi-language samples.

9. **Augment domain structs**
   - Extend `ParsedExpense`, `Expense`, DTOs to include new currency fields.
   - Update constructors and validation.

10. **Enhance CreateExpense use case**
    - Retrieve user profile → `home_currency`.
    - Resolve currency: prefer AI result else fallback.
    - Invoke `ExchangeRateService.Convert` when currencies differ; store rate/home amount.
    - Update repository calls & tests (mock rate service to cover branches).

11. **Update bot responses**
    - Wherever receipts/confirmations are sent (LINE/Telegram handlers), format with `CurrencyAmount` style text `"¥300 (≈NT$64)"`.
    - Add localization strings if needed.

### Phase 3 – HTTP API Layer

12. **Expose currency endpoints**
    - `GET /api/currencies`: query repository, resolve translation based on requester locale.
    - `GET /api/exchange-rates`: optional read-only endpoint for transparency.

13. **User settings endpoints**
    - `GET /api/user/settings`: returns `home_currency`, `locale`.
    - `PUT /api/user/settings`: validate currency code, persist preferences.
    - Ensure auth middleware restricts to owner.

14. **Extend expenses & reports responses**
    - Update serializers to include currency fields.
    - Modify report aggregation query to group by currency and include converted totals.
    - Add unit/integration tests for new response schema (adjust OpenAPI if applicable).

### Phase 4 – Dashboard Frontend (`dashboard/src/app/[locale]/...`)

15. **Create shared CurrencyAmount component**
    - `dashboard/src/components/CurrencyAmount.tsx` receives `amount`, `currency`, `homeAmount`, `homeCurrency`.
    - Handles typography (primary vs secondary) and formatting rules.

16. **User settings page**
    - Add route `/[locale]/user/settings`.
    - Fetch `/api/user/settings` & `/api/currencies` on load, display select inputs.
    - Submit via `PUT /api/user/settings`, show success/error toasts.

17. **Update expenses list & detail**
    - Fetch new fields; render with `CurrencyAmount`.
    - Detail modal shows exchange rate, converted amount; editing original amount triggers PUT and re-fetch.

18. **Update reports page**
    - Display totals in home currency, add table of `by_currency` data with both original and converted totals.
    - Show “updated at” timestamp from `/api/exchange-rates` if endpoint enabled.

19. **Localization**
    - Add i18n strings (e.g., "Home Currency", "≈", "Exchange rate updated").
    - Ensure `/[locale]/...` pages load translations via existing framework.

20. **Frontend tests**
    - Add component tests for `CurrencyAmount`.
    - Add Playwright (or Cypress) scenario: update home currency → list reflects new conversions.

### Phase 5 – Verification & Deployment

21. **Backend testing**
    - Run `make test` (unit + integration) ensuring new modules covered.
    - Add new test cases for migrations (if framework supports) and rate service.

22. **Frontend testing**
    - Run `pnpm test` (or equivalent) for dashboard.
    - Run `pnpm lint` / `pnpm build` ensuring type safety with new props.

23. **Manual QA**
    - Scenario matrix: default home currency, change home currency, input explicit currency, ambiguous input fallback.
    - Verify report totals and API responses through Postman/Thunder Client.

24. **Deployment checklist**
    - Ensure migrations run in production order (1→3).
    - Monitor exchange-rate refresh logs after deploy.
    - Communicate feature change in release notes.
