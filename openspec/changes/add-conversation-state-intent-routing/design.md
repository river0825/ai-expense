## Context
Users can send short, ambiguous messages in Chinese and English. The current backend flow handles report intent with basic keyword checks and otherwise parses expenses. Clarification prompts are not backed by durable pending state, so multi-turn slot filling is unreliable.

## Goals / Non-Goals
- Goals:
  - Support reliable multi-turn clarification for intent workflows.
  - Introduce deterministic intent routing before expense parsing.
  - Preserve all message interactions for analysis.
- Non-Goals:
  - Building a full LLM-only dialogue manager.
  - Implementing every possible intent in v1.

## Decisions
- Decision: Add a persisted conversation state record keyed by user.
  - Why: Supports cross-turn slot filling and timeout handling.
  - Shape: `user_id`, `active_intent`, `pending_slots_json`, `status`, `expires_at`, timestamps.

- Decision: Route intents in a staged pipeline before expense parsing.
  - Why: Keeps latency predictable and avoids accidental expense parsing for command-like messages.
  - Order:
    1) Resume pending state (if active and not expired)
    2) Rule-based intent detection (locale-aware keywords/regex)
    3) Existing parse-conversation flow fallback

- Decision: Keep home currency separate from temporary override.
  - Why: Avoids permanent preference drift when user is traveling.
  - Result: `settings.currency.set` updates `home_currency`; travel intent updates temporary override state.

## Risks / Trade-offs
- Risk: False-positive intent detection on short text.
  - Mitigation: Require strong keyword pattern plus minimum confidence for command intents.

- Risk: Stale pending states cause confusing behavior.
  - Mitigation: Introduce TTL expiry and explicit cancel/reset path.

- Risk: Added complexity in message flow.
  - Mitigation: Keep v1 to a single multi-turn flow (currency clarification) and comprehensive tests.

## Migration Plan
1. Add conversation state table and repository implementation.
2. Add domain interfaces and usecase orchestration hooks.
3. Introduce intent detector with locale-aware rule sets.
4. Implement currency clarification multi-turn flow.
5. Add tests for happy path, timeout, and fallback behavior.

## Open Questions
- Should pending conversation state be shared across all channels for a user ID, or namespaced by channel/session?
- What TTL should be used by default for pending slots (e.g., 10 min vs 30 min)?
