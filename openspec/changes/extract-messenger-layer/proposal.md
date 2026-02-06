# Change: Extract Messenger Layer

## Why

Currently, each messenger integration implements its own orchestration logic, leading to code duplication and maintenance burden. We need a unified layer that handles message processing regardless of the source.

## What Changes

- Extract unified ProcessMessageUseCase that handles orchestration
- Define domain models for messenger abstraction (UserMessage, MessageResponse)
- Refactor all messenger adapters to use the new unified use case
- Maintain backward compatibility for external webhook contracts

## Impact

- Affected specs: line-integration, messenger-gateway (new)
- Improves maintainability by centralizing message processing logic
- Enables easier addition of new messenger platforms
