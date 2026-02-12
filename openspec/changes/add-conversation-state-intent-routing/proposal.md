# Change: Add Multi-turn Conversation State and Intent Routing

## Why
The current message flow treats most user messages as single-turn commands. The bot can ask clarification questions, but it does not persist pending conversational state to reliably interpret the next user message as an answer.

## What Changes
- Add an intent-routing stage before expense parsing to detect operational intents (report, travel mode, currency setting) with locale-aware rules.
- Add persisted pending conversation state so follow-up replies can fill required slots across turns.
- Add a first supported clarification flow for currency switching, including follow-up answer handling and timeout behavior.
- Keep immutable interaction logs for all user messages and bot replies to support future analytics.

## Impact
- Affected specs: `conversation-parsing`, `line-integration`, `expense-management`
- Affected code (expected): `internal/usecase/process_message.go`, new conversation state domain/repository/usecase files, repository implementation, and related tests
- Data model: adds conversation state persistence for pending intent/slots
