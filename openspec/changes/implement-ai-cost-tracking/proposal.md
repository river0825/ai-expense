# Change: Implement Complete AI Cost Tracking with Actual Token Usage and Unit Pricing

## Why

Currently, the system logs AI API usage with costs hardcoded to `0` and token counts estimated from text length. This provides no visibility into actual API spend or ability to enforce budget controls. To properly track and audit AI costs across different providers and models, we need:

1. **Actual token counts**: Extract real token usage from API responses instead of estimating from input text length
2. **Dynamic pricing lookup**: Store provider/model pricing in the database and calculate costs based on actual tokens consumed
3. **Provider flexibility**: Support different pricing for input vs. output tokens (as most LLM providers charge differently)
4. **Extensibility**: Design for easy addition of new AI providers with different pricing models

## What Changes

- **NEW**: Database pricing configuration table to store input/output token prices per provider/model
- **NEW**: Pricing repository to retrieve and manage pricing data
- **NEW**: Extended AI Service responses to include actual token counts from API responses
- **MODIFIED**: AI Service to extract real token metadata from Gemini API responses (currently not captured)
- **MODIFIED**: UseCase layer to perform cost calculation instead of AI Service (separation of concerns)
- **MODIFIED**: Cost logging to store real token counts and calculated costs

## Impact

**Affected Specs:**
- `ai-service`: Token metadata in responses, actual cost calculation

**Affected Code:**
- `internal/domain/models.go` - Add PricingConfig model
- `internal/domain/repositories.go` - Add PricingRepository interface
- `internal/ai/service.go` - Extend Service interface to return token metadata
- `internal/ai/gemini.go` - Extract actual token counts from API responses
- `internal/usecase/` - Add cost calculation logic
- `internal/adapter/repository/` - Implement PricingRepository for SQLite/PostgreSQL
- Database migrations - Create pricing config table

**Breaking Changes:**
- AI Service interface methods now return token metadata (wrapped in new response types)
- Cost is no longer logged within AI Service; UseCase layer now responsible for cost calculation
