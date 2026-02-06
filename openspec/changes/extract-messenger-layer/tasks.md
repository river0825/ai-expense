# Implementation Tasks: Extract Messenger Layer

## Phase 2: Implementation (After Approval)

### Task 5: Define Domain Types for Messenger Layer
- Create `internal/domain/messenger.go`
- Define `UserMessage` struct (UserID, Content, Source, Metadata)
- Define `MessageResponse` struct (Text, Data)

**Acceptance Criteria**:
- [ ] File exists with `UserMessage` and `MessageResponse` structs
- [ ] `go build ./internal/domain` passes

### Task 6: Implement ProcessMessageUseCase (TDD)
- Create `internal/usecase/process_message.go` and `_test.go`
- Implement orchestration: AutoSignup -> Parse -> Create -> Format
- Copy logic from `internal/adapter/messenger/terminal/usecase.go` but adapt to new types

**Acceptance Criteria**:
- [ ] Unit tests cover: Success path, Parse error, Signup error
- [ ] Logic matches existing `TerminalUseCase` exactly
- [ ] `go test ./internal/usecase/...` passes

### Task 7: Refactor Terminal Adapter
- Modify `internal/adapter/messenger/terminal/handler.go`
- Map `TerminalRequest` -> `UserMessage`
- Call `ProcessMessageUseCase`
- Map `MessageResponse` -> HTTP Response
- Delete `internal/adapter/messenger/terminal/usecase.go`

**Acceptance Criteria**:
- [ ] Handler compiles with new UseCase
- [ ] `handler_test.go` updated to mock `ProcessMessageUseCase`
- [ ] `go test ./internal/adapter/messenger/terminal` passes

### Task 8: Refactor Line Adapter
- Modify `internal/adapter/messenger/line/handler.go`
- Map `LineEvent` -> `UserMessage` (Metadata: `ReplyToken`)
- Call `ProcessMessageUseCase`
- Async Handling: On success, use `lineClient.ReplyMessage(token, response.Text)`
- Delete `internal/adapter/messenger/line/usecase.go`

**Acceptance Criteria**:
- [ ] Handler compiles
- [ ] `handler_test.go` passes

### Task 9: Refactor Discord Adapter
- Modify `internal/adapter/messenger/discord/handler.go`
- Map `DiscordInteraction` -> `UserMessage`
- Call `ProcessMessageUseCase`
- Async Handling: On success, use `discordClient.SendMessage` (or interaction callback)
- Delete `internal/adapter/messenger/discord/usecase.go`

**Acceptance Criteria**:
- [ ] Handler compiles
- [ ] `handler_test.go` passes

### Task 10: Refactor Telegram Adapter
- Modify `internal/adapter/messenger/telegram/handler.go`
- Map `TelegramUpdate` -> `UserMessage` (Metadata: `ChatID`)
- Call `ProcessMessageUseCase`
- Async Handling: On success, use `telegramClient.SendMessage`
- Delete `internal/adapter/messenger/telegram/usecase.go`

**Acceptance Criteria**:
- [ ] Handler compiles
- [ ] `handler_test.go` passes

### Task 11: Refactor Remaining Adapters (Slack, Teams, Whatsapp)
- Apply same pattern: Map Input -> UseCase -> Handle Response
- Delete their specific UseCases

**Acceptance Criteria**:
- [ ] All adapters compile
- [ ] All adapter tests pass

### Task 12: Update Dependency Injection (Wiring)
- Modify `cmd/server/main.go` (or wherever dependency injection happens)
- Initialize ONE `ProcessMessageUseCase`
- Inject this single instance into ALL messenger handlers

**Acceptance Criteria**:
- [ ] `go build ./cmd/server` succeeds
- [ ] Server starts up without panics

### Task 13: Verify System
- Run all tests
- Run E2E tests

**Acceptance Criteria**:
- [ ] `go test ./...` passes (All unit tests)
- [ ] `go test ./test/e2e/...` passes
