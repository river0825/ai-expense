## 1. Domain & Repository
- [ ] 1.1 Define `AICostLog` struct in `internal/domain/models.go` (fields: ID, UserID, Operation, Model, InputTokens, OutputTokens, TotalCost, Currency, CreatedAt)
- [ ] 1.2 Define `AICostRepository` interface in `internal/domain/repositories.go`

## 2. Infrastructure (Database)
- [ ] 2.1 Create SQL migration for `ai_cost_logs` table
- [ ] 2.2 Implement `AICostRepository` in `internal/adapter/repository/sqlite`
- [ ] 2.3 Add `AICostRepo` to `Repository` container/struct if applicable

## 3. AI Service Implementation
- [ ] 3.1 Define pricing constants for Gemini (and placeholders for others)
- [ ] 3.2 Update `NewGeminiAI` signature to accept `domain.AICostRepository`
- [ ] 3.3 Update `GeminiAI.ParseExpense` to calculate cost and call `repo.Create`
- [ ] 3.4 Update `GeminiAI.SuggestCategory` to calculate cost and call `repo.Create`
- [ ] 3.5 Update `service.go` factory/initialization logic in `cmd/server` to inject the repository

## 4. Testing
- [ ] 4.1 Update `GeminiAI` tests to mock `AICostRepository`
- [ ] 4.2 Verify cost calculation logic
- [ ] 4.3 Verify persistence in integration tests
