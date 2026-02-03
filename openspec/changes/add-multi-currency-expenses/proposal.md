# Change: Add multi-currency expense tracking

## Why
Users currently have to record every expense in TWD, which blocks accurate tracking while traveling or operating in other currencies. They also cannot configure a default/home currency, so AI parsing assumptions are brittle. We need to persist the original currency, convert to a home currency, and expose that information through APIs and the dashboard.

## What Changes
- introduce currency reference tables, exchange-rate cache, and user preference columns in the database (**BREAKING** migrations)
- extend domain models, AI parsing, and use cases to capture `currency` + `currency_original`, convert via a provider abstraction, and store both original and home amounts
- add HTTP endpoints for currencies and user settings plus updated expense/report payloads
- update the dashboard (`/[locale]/user`) with a currency selector, shared amount component, and reports/list rendering of both values

## Impact
- Affected specs: `expense-management`, `dashboard-metrics`, `reporting` (expense storage + presentation)
- Affected code: migrations, `internal/domain`, `internal/usecase`, `internal/ai`, HTTP handlers, dashboard app, tests, docs
