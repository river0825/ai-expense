## Context
- Clean Architecture Go backend with Gemini-augmented parsing and multi-messenger ingest
- Existing expenses assume TWD and lack user currency preferences
- Dashboard uses Next.js 14 app router with locale segments at `dashboard/src/app/[locale]/`

## Goals
- Persist both original and home-currency amounts per expense
- Provide configurable home currency per user with localization-ready currency names
- Refresh exchange rates daily via Frankfurter API and expose them through a provider/service abstraction
- Update AI prompts, use cases, APIs, and dashboard UI to display and manipulate multi-currency data

## Decisions
1. **Reference Data Tables** — Use normalized `currencies` + `currency_translations` tables plus `exchange_rates` cache per day to avoid depending on live external calls at write time.
2. **Frankfurter Provider** — Implement an `ExchangeRateProvider` interface with an initial Frankfurter-backed adapter (EUR default, `base=` overrides) to keep swapping providers trivial.
3. **Dual-Amount Storage** — Store `original_amount` + `currency`, along with `home_amount`, `home_currency`, and `exchange_rate` to allow deterministic historical reporting without re-fetching rates.
4. **AI Prompt Update** — Extend Gemini prompt schema to include `currency` and `currency_original`, falling back to the user's home currency upon ambiguity.
5. **UI Rendering** — Introduce a shared `CurrencyAmount` component to ensure consistent formatting (primary original amount, secondary converted amount) across expense list/detail/report contexts.

## Risks / Mitigations
- **Frankfurter downtime** — Cache latest successful rate set and log warnings; allow manual refresh via admin endpoint if needed.
- **Missing rate for target date** — Fallback to last available business day; surface rate date in responses for transparency.
- **AI misclassification** — Persist `currency_original` for debugging and possible correction flows; default to home currency when unspecified.
- **Breaking migrations** — Ship migrations in stages (reference tables → user preferences → expense expansion) with careful backfill scripts.

## Migration Plan
1. Add `currencies`, `currency_translations`, and `exchange_rates` tables with seed data via migration 1.
2. Add `home_currency` + `locale` columns to `users` with defaults/backfill via migration 2.
3. Rename `amount` → `original_amount` and add new currency columns to `expenses`, backfilling existing rows via migration 3.
4. Deploy code that reads/writes the new fields after migrations land; enable Frankfurter refresh job.

## References
- Detailed implementation breakdown: `docs/plans/2026-01-29-multi-currency-design.md`
