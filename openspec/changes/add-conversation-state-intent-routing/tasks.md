## 1. Implementation
- [ ] 1.1 Add conversation state domain model and repository interface for active intent and pending slots.
- [ ] 1.2 Add persistence schema/repository implementation for conversation state with expiry support.
- [ ] 1.3 Add intent detector module with locale-aware rule sets for report, travel mode, and currency setting intents.
- [ ] 1.4 Update `ProcessMessageUseCase` flow to resume pending state first, then detect intent, then fallback to parse expense.
- [ ] 1.5 Implement multi-turn currency clarification flow: ask for currency when ambiguous and consume next user answer.
- [ ] 1.6 Ensure every incoming/outgoing interaction remains logged for analytics.

## 2. Verification
- [ ] 2.1 Add unit tests for intent detection in zh-TW and en-US examples.
- [ ] 2.2 Add unit tests for pending-slot round-trip (ask currency -> user answers -> update applied).
- [ ] 2.3 Add unit tests for pending-state expiry and cancel/reset behavior.
- [ ] 2.4 Run `go test ./...` and confirm no regressions.
