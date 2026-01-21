## Context
We need to track the financial cost of using AI APIs. Different providers have different pricing models (usually per 1k or 1M tokens, separated by input/output).

## Decisions
- **Decision**: Handle cost tracking inside the `AIService` implementation (Adapter layer).
- **Why**: 
  - Keeps the Use Case layer clean (Business logic doesn't care about cost tracking).
  - The Adapter is the only place that knows the specific model version and token usage returned by the API response.
  - Allows "fire and forget" or async logging without blocking the main response if needed (though synchronous is safer for data integrity initially).

- **Decision**: Store costs in a local SQLite table `ai_cost_logs`.
- **Why**: Consistent with current architecture. Allows easy querying for future reporting dashboards.

## Models
### AICostLog
- `id`: UUID
- `user_id`: String (to track cost per user)
- `operation`: String ("parse_expense", "suggest_category")
- `provider`: String ("gemini", "openai")
- `model`: String ("gemini-2.5-flash", "gpt-4o")
- `input_tokens`: Int
- `output_tokens`: Int
- `cost`: Float64 (standardized to USD or local currency config)
- `currency`: String ("USD")
- `created_at`: Timestamp

## Open Questions
- **Pricing Configuration**: Should prices be hardcoded or configurable? 
  - *Answer*: Hardcode current list prices in a `pricing.go` helper within `internal/ai` for now. Move to config later if frequent changes occur.

## Risks
- **Performance**: Database write on every AI call adds latency.
  - *Mitigation*: It's negligible compared to the LLM latency (hundreds of ms vs <10ms for SQLite).
