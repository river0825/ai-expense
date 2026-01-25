# Design: AI Cost Tracking Architecture

## Context

The system currently logs AI API usage but with hardcoded costs ($0) and estimated token counts. To enable accurate cost tracking and budget management, we need to:
1. Capture actual token counts from API responses
2. Store pricing data for different providers/models
3. Calculate costs dynamically in the business logic layer

## Goals

- **Accurate Cost Tracking**: Use real token counts from APIs instead of estimates
- **Provider Flexibility**: Support multiple AI providers with different pricing structures
- **Separation of Concerns**: Cost calculation belongs in UseCase, not AI Service
- **Extensibility**: Easy to add new providers and pricing models without code duplication
- **Auditability**: Full record of costs with provider, model, tokens, and calculated price

## Non-Goals

- Real-time pricing updates from provider APIs (static config for now)
- Multi-currency conversion (USD only; can add later)
- Budget enforcement (post-tracking alerting; future feature)

## Decisions

### 1. Pricing Storage (Database Table)
**Decision**: Store pricing in database table `ai_pricing_config`

**Rationale**:
- Supports dynamic pricing updates without code deployment
- Easy to version and track pricing changes over time
- Can be managed via admin UI in future

**Structure**:
```
ai_pricing_config:
  id (PK)
  provider (e.g., "gemini", "claude", "openai")
  model (e.g., "gemini-2.5-lite", "claude-3-opus")
  input_token_price (USD per 1M tokens)
  output_token_price (USD per 1M tokens)
  currency ("USD")
  effective_date (when pricing becomes active)
  is_active (boolean)
  created_at, updated_at
```

### 2. Token Metadata in Responses
**Decision**: Extend response structs to include token metadata

**Rationale**:
- AI Service knows actual token usage from API responses
- Usecase needs this info to calculate costs
- Cleaner than multiple API calls

**New Response Types**:
```go
type ParseExpenseResponse struct {
  Expenses []*ParsedExpense
  Tokens   *TokenMetadata  // New
}

type TokenMetadata struct {
  InputTokens  int
  OutputTokens int
  TotalTokens  int
}
```

### 2b. Cost Note Field
**Decision**: Add optional `cost_note` field to `AICostLog` for special conditions

**Rationale**:
- Enables audit trail for missing pricing, fallbacks, cache hits
- Helps diagnose zero-cost entries
- Examples: "pricing_not_configured", "fallback_parsing", "cache_hit"

### 3. Cost Calculation Location
**Decision**: Calculate costs in UseCase layer, not AI Service

**Rationale**:
- AI Service responsible for parsing, not accounting
- UseCase orchestrates: AI call → pricing lookup → cost calculation → persist
- Easier to test and refactor
- Supports multiple AI services in same flow

**Flow**:
```
UseCase.ParseMessage()
  ├─ Call AIService.ParseExpense() → get expenses + token metadata
  ├─ Look up pricing from PricingRepository
  ├─ Calculate cost = (inputTokens * inputPrice + outputTokens * outputPrice) / 1,000,000
  ├─ Create AICostLog with real tokens and cost
  └─ Persist via AICostRepository
```

### 4. Gemini API Token Extraction
**Decision**: Parse `usageMetadata` from Gemini API response

**Current**: Estimating tokens from input text length and hardcoding 50 output tokens
**New**: Extract from API response `usageMetadata.prompt_token_count` and `usageMetadata.candidates[0].content.usage_metadata.output_token_count`

**Impact**: More accurate token counts, though Gemini API docs need verification for exact field names.

### 5. Extensibility for Other Providers
**Design Pattern**: Template for adding new providers
- Each provider (Claude, OpenAI) implements same response wrapper
- Pricing table supports any provider/model combo
- No code changes needed after pricing is added to database

## Alternatives Considered

### Alternative 1: Cost calculation in AI Service
**Pros**: Encapsulated accounting logic
**Cons**: AI Service becomes responsible for pricing lookup (violation of SRP); harder to test independently
**Rejected**: Violates single responsibility principle

### Alternative 2: Hardcode pricing in provider implementations
**Pros**: Simple, localized to each provider
**Cons**: Duplicates pricing logic; requires code change for pricing updates; breaks separation of concerns
**Rejected**: Not extensible; difficult to maintain

### Alternative 3: Call external pricing API
**Pros**: Always up-to-date pricing
**Cons**: Adds external dependency; latency; complexity; overkill for static pricing
**Rejected**: Out of scope; database table sufficient for now

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Gemini API response format changes | Document expected format; add unit tests verifying token extraction |
| Pricing outdated | Add admin UI for pricing updates; log warnings when pricing not found |
| Cost calculation errors | Unit tests for calculation logic; spot-check against provider bills |
| New provider added but no pricing configured | Graceful fallback: log warning, save cost as 0, continue processing |

## Migration Plan

1. **Phase 1**: Add pricing table and repository (backward compatible)
2. **Phase 2**: Modify AI Service to return token metadata (breaking change to interface)
3. **Phase 3**: Update UseCase to calculate costs (where actual savings come from)
4. **Phase 4**: Update Gemini implementation to extract real tokens from API
5. **Phase 5**: Deprecate old cost logging in AI Service; clean up gradual

No data migration needed (old logs remain as-is; new logs use real costs).

## Open Questions

1. **Gemini API token field names**: Confirm exact field paths in `usageMetadata` from actual API responses
2. **Pricing seed data**: Should migration include seed data for current Gemini pricing, or manage separately?
3. **Retroactive cost calculation**: Should we backfill historical logs with calculated costs? (Probably not; too complex)
