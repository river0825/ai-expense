# Messenger Layer Abstraction Plan

## Context

### Original Request
Extract a layer for different sources of incoming requests (messengers). The layer is responsible for transforming different webhook schemas to a use case that handles actual logic.

### User Request (Update)
The user explicitly requested to create an **OpenSpec Change Proposal** for this work.

### Interview Summary
**Key Discussions**:
- **Scope**: Refactor ALL messengers (Terminal, Line, Discord, Telegram, etc.) to use the new layer.
- **Response Strategy**: The Use Case returns a standard `Response` object; the Adapter handles delivery (sync write vs async API).
- **Architecture**: Move from N UseCases to 1 `ProcessMessageUseCase` + N Adapters.

**Research Findings**:
- **Fragmentation**: Currently each messenger has a dedicated UseCase duplicating orchestration logic.
- **Protocol Differences**:
  - `Terminal`: Sync HTTP response.
  - `Line`/`Discord`/`Telegram`: Async API calls (sometimes triggering immediately after webhook).
- **Test Infrastructure**: Exists (`handler_test.go`, `e2e` tests). We can leverage this.

### Metis Review (Self-Conducted)
**Identified Gaps (Addressed)**:
- **Rich Media**: Explicitly OUT of scope. Text only.
- **Dependency Injection**: Need to update `cmd/server/main.go` to wire the new UseCase.
- **Regression**: Must ensure `webhook_flow_test.go` passes.

---

## Work Objectives

### Core Objective
1. Formalize the change via OpenSpec Proposal.
2. Replace duplicated messenger logic with a single `ProcessMessageUseCase` and lightweight Adapters.

### Concrete Deliverables
- **OpenSpec**: `openspec/changes/extract-messenger-layer/` (Proposal, Design, Specs).
- **Code**: `internal/usecase/process_message.go`, `internal/domain/messenger.go`.
- **Refactor**: All messenger adapters in `internal/adapter/messenger/*`.

### Definition of Done
- [ ] OpenSpec proposal validated (`openspec validate extract-messenger-layer --strict`).
- [ ] New `ProcessMessageUseCase` handles the orchestration.
- [ ] All messenger adapters compile and pass their unit tests.
- [ ] E2E tests `test/e2e/webhook_flow_test.go` pass.

### Must Have
- Unified `ProcessMessageUseCase`.
- Backward compatibility for external webhook contracts (URLs/Signatures).
- Exact text response matching.

### Must NOT Have
- Support for images/video/location.
- New features (Edit/Delete).

---

## Task Flow

```
Phase 1: OpenSpec Proposal → Phase 2: Implementation
```

---

## TODOs

### Phase 1: OpenSpec Proposal

- [ ] 1. Scaffold OpenSpec Change
  **What to do**:
  - Create directory: `openspec/changes/extract-messenger-layer/`
  - Create subdirs: `specs/messenger-gateway` and `specs/line-integration`
  - Create `proposal.md`:
    ```markdown
    # Change: Extract Messenger Layer
    ## Why
    Currently, each messenger integration implements its own orchestration logic...
    ## What Changes
    - Extract unified ProcessMessageUseCase.
    - Define domain models.
    - Refactor adapters.
    ## Impact
    - Affected specs: line-integration, messenger-gateway (new)
    ```
  - Create `tasks.md` (Copy relevant implementation tasks from Phase 2 below).

- [ ] 2. Define New Capability Spec (Messenger Gateway)
  **What to do**:
  - Create `openspec/changes/extract-messenger-layer/specs/messenger-gateway/spec.md`.
  - Content:
    ```markdown
    ## ADDED Requirements
    ### Requirement: Unified Message Processing
    The system SHALL provide a unified gateway to process messages from any supported messenger source.

    #### Scenario: Text Message Processing
    - **WHEN** a text message is received from any source
    - **THEN** system attempts to parse expenses
    - **AND** system creates expenses if valid
    - **AND** system generates a standardized text response
    ```

- [ ] 3. Modify Existing Spec (Line Integration)
  **What to do**:
  - Create `openspec/changes/extract-messenger-layer/specs/line-integration/spec.md`.
  - Content:
    ```markdown
    ## MODIFIED Requirements
    ### Requirement: Receive Messages from LINE Messaging API
    The system SHALL receive incoming messages from LINE and delegate processing to the Messenger Gateway.

    #### Scenario: Webhook Delegation
    - **WHEN** LINE webhook is received
    - **THEN** system transforms event to UserMessage
    - **AND** delegates logic to Messenger Gateway
    ```
    (Note: Copy full requirement text if modifying behavior significantly, but here we are mainly changing internal delegation. If behavior is preserved, we might not strictly need MODIFIED, but it's good practice to reflect the architectural shift).

- [ ] 4. Validate Proposal
  **What to do**:
  - Run `openspec validate extract-messenger-layer --strict`.
  - Fix any issues.

### Phase 2: Implementation (After Approval)

- [ ] 5. Define Domain Types for Messenger Layer
  **What to do**:
  - Create `internal/domain/messenger.go`.
  - Define `UserMessage` struct (UserID, Content, Source, Metadata).
  - Define `MessageResponse` struct (Text, Data).
  
  **Acceptance Criteria**:
  - [ ] File exists with `UserMessage` and `MessageResponse` structs.
  - [ ] `go build ./internal/domain` passes.

- [ ] 6. Implement ProcessMessageUseCase (TDD)
  **What to do**:
  - Create `internal/usecase/process_message.go` and `_test.go`.
  - Implement orchestration: AutoSignup -> Parse -> Create -> Format.
  - **Important**: Copy logic from `internal/adapter/messenger/terminal/usecase.go` but adapt to new types.
  
  **Acceptance Criteria**:
  - [ ] Unit tests cover: Success path, Parse error, Signup error.
  - [ ] Logic matches existing `TerminalUseCase` exactly.
  - [ ] `go test ./internal/usecase/...` passes.

- [ ] 7. Refactor Terminal Adapter
  **What to do**:
  - Modify `internal/adapter/messenger/terminal/handler.go`.
  - Map `TerminalRequest` -> `UserMessage`.
  - Call `ProcessMessageUseCase`.
  - Map `MessageResponse` -> HTTP Response.
  - **Delete** `internal/adapter/messenger/terminal/usecase.go` (it's now obsolete).
  
  **Acceptance Criteria**:
  - [ ] Handler compiles with new UseCase.
  - [ ] `handler_test.go` updated to mock `ProcessMessageUseCase`.
  - [ ] `go test ./internal/adapter/messenger/terminal` passes.

- [ ] 8. Refactor Line Adapter
  **What to do**:
  - Modify `internal/adapter/messenger/line/handler.go`.
  - Map `LineEvent` -> `UserMessage` (Metadata: `ReplyToken`).
  - Call `ProcessMessageUseCase`.
  - **Async Handling**: On success, use `lineClient.ReplyMessage(token, response.Text)`.
  - **Delete** `internal/adapter/messenger/line/usecase.go`.

  **Acceptance Criteria**:
  - [ ] Handler compiles.
  - [ ] `handler_test.go` passes.

- [ ] 9. Refactor Discord Adapter
  **What to do**:
  - Modify `internal/adapter/messenger/discord/handler.go`.
  - Map `DiscordInteraction` -> `UserMessage`.
  - Call `ProcessMessageUseCase`.
  - **Async Handling**: On success, use `discordClient.SendMessage` (or interaction callback).
  - **Delete** `internal/adapter/messenger/discord/usecase.go`.

  **Acceptance Criteria**:
  - [ ] Handler compiles.
  - [ ] `handler_test.go` passes.

- [ ] 10. Refactor Telegram Adapter
  **What to do**:
  - Modify `internal/adapter/messenger/telegram/handler.go`.
  - Map `TelegramUpdate` -> `UserMessage` (Metadata: `ChatID`).
  - Call `ProcessMessageUseCase`.
  - **Async Handling**: On success, use `telegramClient.SendMessage`.
  - **Delete** `internal/adapter/messenger/telegram/usecase.go`.

  **Acceptance Criteria**:
  - [ ] Handler compiles.
  - [ ] `handler_test.go` passes.

- [ ] 11. Refactor Remaining Adapters (Slack, Teams, Whatsapp)
  **What to do**:
  - Apply same pattern: Map Input -> UseCase -> Handle Response.
  - Delete their specific UseCases.
  
  **Acceptance Criteria**:
  - [ ] All adapters compile.
  - [ ] All adapter tests pass.

- [ ] 12. Update Dependency Injection (Wiring)
  **What to do**:
  - Modify `cmd/server/main.go` (or wherever dependency injection happens).
  - Initialize ONE `ProcessMessageUseCase`.
  - Inject this single instance into ALL messenger handlers.
  
  **Acceptance Criteria**:
  - [ ] `go build ./cmd/server` succeeds.
  - [ ] Server starts up without panics.

- [ ] 13. Verify System
  **What to do**:
  - Run all tests.
  - Run E2E tests.
  
  **Acceptance Criteria**:
  - [ ] `go test ./...` passes (All unit tests).
  - [ ] `go test ./test/e2e/...` passes.

---

## Success Criteria

### Verification Commands
```bash
go test ./... -v
go test ./test/e2e/... -v
```

### Final Checklist
- [ ] No more `*UseCase` inside `internal/adapter/messenger/*` (except maybe for specifics not covered).
- [ ] Single `ProcessMessageUseCase` in `internal/usecase`.
- [ ] All 7 messengers working via new layer.
