# Change: Extract Messenger Layer

## Why
Currently, each messenger integration (Line, Terminal, Discord, etc.) implements its own orchestration logic (Signup -> Parse -> Create -> Respond). This leads to code duplication, inconsistent behavior across platforms, and higher maintenance cost. Adding a new messenger requires copying the entire flow.

## What Changes
- **Extract** a unified `ProcessMessageUseCase` (Messenger Gateway) that handles the common orchestration logic.
- **Define** generic `UserMessage` and `MessageResponse` structures in the Domain layer.
- **Refactor** existing `line-integration` and `terminal` handlers to act as lightweight adapters that map platform-specific schemas to the unified domain models.
- **Retain** platform-specific webhook verification and response delivery mechanisms (Sync vs Async) in the adapters.

## Impact
- **Affected Specs**:
  - `line-integration`: Will delegate processing logic to the new gateway.
  - `messenger-gateway` (New): Will define the core orchestration requirements.
- **Affected Code**:
  - `internal/usecase`: New `process_message.go`.
  - `internal/adapter/messenger/*`: All handlers refactored.
  - `cmd/server`: Wiring updated.
