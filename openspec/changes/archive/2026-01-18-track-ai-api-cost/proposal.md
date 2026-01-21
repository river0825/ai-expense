# Change: Track AI API Cost

## Why
Currently, we have no visibility into the cost of AI operations. As we scale or switch to paid providers (like OpenAI), tracking expenses becomes critical to prevent budget overruns and understand per-user unit economics.

## What Changes
- Add `AICostLog` domain model to track token usage and estimated cost per request.
- Add `AICostRepository` to persist cost logs.
- Update `GeminiAI` (and future providers) to calculate and save cost logs automatically after each request.
- Create a new database table `ai_cost_logs`.

## Impact
- **Specs**: `ai-service` (Modified)
- **Code**: 
  - `internal/domain`: New model and repository interface.
  - `internal/ai`: Service implementation update to inject repository and log costs.
  - `internal/adapter/repository`: New SQLite implementation.
  - `migrations`: New table.
