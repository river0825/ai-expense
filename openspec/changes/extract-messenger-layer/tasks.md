## 1. Domain Implementation
- [ ] 1.1 Define `UserMessage` and `MessageResponse` structs in `internal/domain/messenger.go`.

## 2. UseCase Implementation
- [ ] 2.1 Implement `ProcessMessageUseCase` in `internal/usecase/process_message.go` with TDD (`process_message_test.go`).
- [ ] 2.2 Orchestrate AutoSignup -> Parse -> Create -> Format flow.

## 3. Adapter Refactoring
- [ ] 3.1 Refactor `terminal` adapter to use `ProcessMessageUseCase` (map request/response).
- [ ] 3.2 Refactor `line` adapter to use `ProcessMessageUseCase` (map event/response).
- [ ] 3.3 Refactor `discord` adapter to use `ProcessMessageUseCase` (map interaction/response).
- [ ] 3.4 Refactor `telegram` adapter to use `ProcessMessageUseCase` (map update/response).
- [ ] 3.5 Refactor `slack`, `teams`, `whatsapp` adapters.
- [ ] 3.6 Delete obsolete specific UseCases (`TerminalUseCase`, `LineUseCase`, etc.).

## 4. Wiring & Verification
- [ ] 4.1 Update `cmd/server/main.go` to inject the single `ProcessMessageUseCase`.
- [ ] 4.2 Verify all unit tests pass (`go test ./...`).
- [ ] 4.3 Verify E2E tests pass (`test/e2e/webhook_flow_test.go`).
