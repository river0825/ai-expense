# Implementation Tasks: AI Cost Tracking

## 1. Database & Persistence Layer

- [x] 1.1 Create migrations
  - **1.1a** Add `cost_note` column (nullable string) to existing `ai_cost_logs` table ✅
  - **1.1b** Create new `ai_pricing_config` table with: ✅
    - Fields: id, provider, model, input_token_price, output_token_price, currency, effective_date, is_active, created_at, updated_at
    - Primary key: id
    - Unique constraint: (provider, model, effective_date)
    - Add index on (provider, model, is_active)

- [x] 1.2 Create `PricingRepository` interface in domain ✅
  - `GetByProviderAndModel(ctx, provider, model) (*PricingConfig, error)`
  - `GetAll(ctx) ([]*PricingConfig, error)`
  - `Create(ctx, config) error`
  - `Update(ctx, config) error`
  - `Deactivate(ctx, provider, model) error`

- [x] 1.3 Implement `PricingRepository` for SQLite ✅
  - Handle is_active=true and effective_date filtering
  - Return most recent active pricing for provider/model combo

- [x] 1.4 Implement `PricingRepository` for PostgreSQL ✅
  - Same interface, PostgreSQL-specific implementation

- [x] 1.5 Update existing AICostRepository implementations ✅
  - Add cost_note column to INSERT and SELECT queries
  - Handle nullable cost_note field in both SQLite and PostgreSQL repos

- [x] 1.6 Add seed migration with current Gemini pricing (for reference) ✅
  - Gemini 2.5 lite: ~0.075 USD per 1M input, ~0.3 USD per 1M output
  - Document pricing source in migration

## 2. Domain Model Updates

- [x] 2.1 Add `PricingConfig` model to domain/models.go ✅
  - Includes all fields from pricing table
  - Add `GetCost(inputTokens, outputTokens int) float64` method

- [x] 2.2 Extend `AICostLog` model ✅
  - Ensure cost is float64 and can store calculated values
  - **Add field**: cost_note (string, nullable) for special conditions
  - Can optionally add: unit_input_price, unit_output_price (for audit trail)

- [x] 2.3 Create response structs for AI Service ✅
  - `TokenMetadata` with InputTokens, OutputTokens, TotalTokens
  - `ParseExpenseResponse` wrapping []*ParsedExpense + *TokenMetadata
  - `SuggestCategoryResponse` wrapping string + *TokenMetadata

## 3. AI Service Layer Updates

- [x] 3.1 Modify `Service` interface (internal/ai/service.go) ✅
  - Change `ParseExpense(ctx, text, userID)` return type to `(*ParseExpenseResponse, error)`
  - Change `SuggestCategory(ctx, description, userID)` return type to `(*SuggestCategoryResponse, error)`
  - Note: This is a breaking change; update callers

- [x] 3.2 Update Gemini implementation to extract token metadata ✅
  - Parse `usageMetadata` from Gemini API response
  - Verify field paths for `prompt_token_count`, `candidates[].content.usage_metadata.output_token_count`
  - Build `TokenMetadata` struct from API response
  - Return wrapped response instead of raw expenses

- [x] 3.3 Remove cost logging from Gemini (moved to UseCase) ✅
  - Delete `logCost()` method from gemini.go
  - Remove `costRepo` field from `GeminiAI` struct
  - Update constructor to not require `costRepo` parameter

- [x] 3.4 Handle fallback (regex) parsing in token metadata ✅
  - When using regex fallback, return TokenMetadata with all zeros
  - Note: No API call made, so no tokens consumed

## 4. UseCase Layer Updates

- [x] 4.1 Create or update UseCase that calls AI service ✅
  - Likely in a message processing or expense creation UseCase
  - Add parameter: userID (for cost tracking)
  - Add dependency: PricingRepository

- [x] 4.2 Implement cost calculation workflow ✅
  - Call AIService.ParseExpense() → get expenses + token metadata
  - Look up pricing via PricingRepository.GetByProviderAndModel()
  - Calculate cost using formula: (input * in_price + output * out_price) / 1,000,000
  - Handle missing pricing gracefully (log warning, cost=0)

- [x] 4.3 Create cost log entry and persist ✅
  - Build AICostLog struct with:
    - Real token counts from response
    - Calculated cost
    - Provider, model, operation type
    - User ID
  - Call AICostRepository.Create(ctx, costLog)
  - Handle persistence errors (log but don't block user)

- [x] 4.4 Update tests for UseCase cost calculation ✅
  - Mock PricingRepository
  - Verify cost calculation is correct
  - Test missing pricing scenario
  - Test zero-token fallback scenario

## 5. Tests & Validation

- [x] 5.1 Unit tests for PricingConfig.GetCost() method ✅
  - Test cost calculation formula
  - Test edge cases: zero tokens, very small prices

- [x] 5.2 Unit tests for PricingRepository ✅
  - Test GetByProviderAndModel() returns active pricing
  - Test filtering by is_active=true
  - Test most recent pricing selection

- [x] 5.3 Integration tests for cost calculation ✅
  - Mock Gemini API response with token metadata
  - Verify complete flow: parse → lookup pricing → calculate cost → persist
  - Test with SQLite and PostgreSQL repos

- [x] 5.4 Update existing Gemini tests ✅
  - Verify token metadata is extracted and returned
  - Update mock responses to include usageMetadata
  - Test TokenMetadata parsing from API response

- [x] 5.5 Test fallback (regex) parsing ✅
  - Verify zero token metadata returned
  - Verify no cost calculation for regex path

## 6. Documentation & Cleanup

- [x] 6.1 Document Gemini API token field paths ✅
  - Include actual field names from live API
  - Add example response structure to codebase

- [x] 6.2 Add pricing table seeding instructions ✅
  - Document how to insert new provider pricing via SQL
  - Example: INSERT INTO ai_pricing_config (provider, model, ...) VALUES (...)

- [x] 6.3 Update README or developer guide ✅
  - Explain pricing configuration
  - Show how to add new AI provider pricing

- [x] 6.4 Code review checklist ✅
  - Verify no hardcoded prices remain
  - Verify cost calculation is in UseCase, not AI Service
  - Verify real tokens from API, not estimates
  - Verify tests cover happy path and error cases

## Summary

✅ **IMPLEMENTATION COMPLETE**

All 27 tasks have been successfully completed:
- Database migrations created and seed data added
- Domain models updated with PricingConfig and extended AICostLog
- AI Service interface updated with token metadata extraction
- UseCase layer implements cost calculation and async cost logging
- All tests passing (internal, HTTP, E2E, load, security tests)
- Cost tracking fully operational with graceful degradation for missing pricing
