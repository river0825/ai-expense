# Pricing Sync Feature Design

**Date:** 2025-01-25
**Status:** Approved
**Provider Abstraction:** Yes

## Overview

On-demand pricing sync feature with provider abstraction. Admins trigger sync via API endpoint to fetch latest pricing from any supported provider (Gemini, Claude, etc.). Uses append-only audit trail—only inserts new rows when prices change.

## Architecture

### Core Components

1. **PricingProvider Interface** (Domain Layer)
   - Contract for fetching pricing from any provider
   - Methods: `Fetch(ctx)`, `Provider()`

2. **GeminiPricingProvider** (AI Service Layer)
   - Implements PricingProvider
   - Scrapes https://ai.google.dev/pricing
   - Extracts all Gemini models with prices
   - Retries 3 times with exponential backoff (1s, 2s, 4s)
   - Logs each attempt

3. **PricingSyncUseCase** (UseCase Layer)
   - Generic, provider-agnostic orchestrator
   - Compares fetched prices vs. current active pricing
   - Only inserts new rows when prices change
   - Deactivates old rows before inserting new

4. **PricingHandler** (HTTP Adapter Layer)
   - Admin authentication via X-API-Key
   - Routes to providers via query parameter
   - 5 endpoints: sync, list, create, update, delete

### API Endpoints

All endpoints require `X-API-Key` header.

**POST /api/pricing/sync?provider=gemini**
- Trigger sync for specified provider
- Response: `{success: true, provider: "gemini", synced_at: "...", models_updated: 5, models_unchanged: 2, errors: []}`
- On failure: `{success: false, error: "..."}`

**GET /api/pricing**
- List all pricing configs (active + historical)
- Optional: `?active=true` to show only active rows

**POST /api/pricing**
- Create manual pricing entry
- Body: `{provider, model, input_token_price, output_token_price}`

**PUT /api/pricing/{id}**
- Update existing pricing config

**DELETE /api/pricing/{id}**
- Deactivate pricing config

## Data Flow: Sync Process

1. Admin hits `POST /api/pricing/sync?provider=gemini`
2. Handler authenticates, invokes PricingSyncUseCase.Sync()
3. UseCase calls GeminiPricingProvider.Fetch() (with retries)
4. Fetcher scrapes Google pricing page, returns `[]*PricingConfig`
5. UseCase compares each model:
   - If no current active pricing: insert new row
   - If current pricing differs: deactivate old, insert new
   - If unchanged: skip
6. Return result with counts of updated/unchanged models

## Append-Only Audit Trail

Every pricing change creates a permanent record:

```sql
-- Before: old pricing active
SELECT * FROM ai_pricing_config
WHERE provider='gemini' AND model='gemini-2.5-lite' AND is_active=true;
-- Result: id='pricing_gemini_2.5_lite_2025_01_25', input_price=0.075, is_active=true

-- Sync fetches new price: 0.076
-- Step 1: Deactivate old
UPDATE ai_pricing_config SET is_active=false
WHERE provider='gemini' AND model='gemini-2.5-lite' AND is_active=true;

-- Step 2: Insert new
INSERT INTO ai_pricing_config (..., is_active=true, effective_date='2025-01-25 10:30:00');
-- New id='pricing_gemini_2.5_lite_2025_01_25_10_30_00', input_price=0.076
```

History preserved: both rows exist, old is inactive, new is active.

## Error Handling & Resilience

**Fetch Failures:**
- Retry up to 3 times with exponential backoff
- Log each attempt: `[WARN] pricing_fetch attempt X/3 failed: {error}`
- Final failure logs: `[ERROR] pricing_fetch failed after 3 attempts`

**Graceful Degradation:**
- If sync fails: existing pricing remains, no partial updates
- System continues operating with old pricing
- Error logged and returned to admin

**No Corrupted State:**
- Deactivate-before-insert pattern prevents orphans
- If insert fails: old pricing remains active (safe fallback)
- Database constraints ensure uniqueness

**Logging:**
- Start: `Sync started for provider=gemini`
- Progress: `Found 7 models in pricing page`
- Changes: `Model gemini-2.5-lite: price changed (was 0.075, now 0.076)`
- Unchanged: `Model gemini-2.0-flash: price unchanged (0.30)`
- Completion: `Sync completed: 3 updated, 4 unchanged, 0 errors`

## Testing Strategy

**Unit Tests (pricing_provider_test.go):**
- Success case with mock HTTP response
- Retry on network error (fail, fail, succeed)
- All retries fail
- Invalid/malformed HTML
- Parse errors

**Unit Tests (pricing_sync_test.go):**
- All new prices (insert)
- Some prices changed (deactivate + insert)
- No prices changed (skip)
- Missing current pricing (insert)
- Fetch fails (error returned, repo untouched)

**Integration Tests (pricing_handler_test.go):**
- POST /api/pricing/sync success and failure cases
- GET /api/pricing listing
- POST /api/pricing create
- PUT /api/pricing update
- DELETE /api/pricing delete
- Authentication (missing key, invalid key)

**Test Mocks:**
- MockPricingProvider with hardcoded models
- MockPricingRepository for in-memory testing

## Implementation Phases

**Phase 1: Domain Layer**
- Add `PricingProvider` interface
- Verify `PricingRepository` complete

**Phase 2: AI Service Layer**
- Create GeminiPricingProvider with HTML scraping
- Implement retry logic with logging
- Unit tests

**Phase 3: UseCase Layer**
- Create PricingSyncUseCase
- Implement compare/deactivate/insert logic
- Unit tests

**Phase 4: HTTP Adapter Layer**
- Create PricingHandler
- Implement 5 endpoints
- Integration tests

**Phase 5: Server Integration**
- Instantiate providers and use case
- Register routes
- Manual testing

**Phase 6: Dependencies**
- Add `github.com/PuerkitoBio/goquery` to go.mod

## Extensibility

To add support for a new provider (Claude, OpenAI, etc.):

1. Create new struct `ClaudePricingProvider` implementing `PricingProvider`
2. Implement `Fetch()` with provider-specific scraping/parsing
3. Register in PricingHandler.providers map
4. No changes to UseCase or core logic needed

Provider abstraction allows unlimited scaling to multiple AI providers.
