## 1. Database & Domain
- [ ] 1.1 Add currency, currency_translations, and exchange_rates tables with seed data
- [ ] 1.2 Add home_currency + locale columns to users
- [ ] 1.3 Expand expenses table with currency/home fields and backfill existing data
- [ ] 1.4 Introduce Currency + ExchangeRate domain models and repositories
- [ ] 1.5 Add ExchangeRateProvider abstraction with Frankfurter implementation
- [ ] 1.6 Build ExchangeRateService and daily refresh job

## 2. AI & Use Cases
- [ ] 2.1 Update Gemini prompts + parser to emit currency + currency_original
- [ ] 2.2 Extend ParsedExpense, Expense, and DTOs with currency fields
- [ ] 2.3 Update CreateExpense and ProcessMessage flows to resolve currencies and convert via service
- [ ] 2.4 Update bot responses to show original + converted amounts

## 3. API Layer
- [ ] 3.1 Add /api/currencies endpoint returning localized currency names
- [ ] 3.2 Add GET/PUT /api/user/settings for home_currency + locale
- [ ] 3.3 Update /api/expenses + /api/reports responses with currency data
- [ ] 3.4 (Optional) Add /api/exchange-rates for transparency/debugging

## 4. Dashboard
- [ ] 4.1 Add CurrencyAmount component and supporting utilities
- [ ] 4.2 Build `/[locale]/user/settings` page for home currency selection
- [ ] 4.3 Update expense list/detail views to show both currency values
- [ ] 4.4 Update reports UI with by-currency breakdown + refresh timestamp
- [ ] 4.5 Add localization strings/tests for new UI elements

## 5. Verification
- [ ] 5.1 Extend backend/unit tests + run `go test ./...`
- [ ] 5.2 Extend dashboard tests + run `bun run lint && bun run test`
- [ ] 5.3 Manual QA scenarios + update release notes
