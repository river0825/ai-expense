# Change: Enable messengers via environment variable with terminal as default

## Why
Currently, the server requires `LINE_CHANNEL_TOKEN` to start, making local development and testing cumbersome. Developers need LINE credentials even when they just want to test the core expense tracking logic. The terminal messenger already exists but isn't wired into the server startup.

## What Changes
- Add `ENABLED_MESSENGERS` environment variable to control which messengers are active
- Default to `terminal` messenger when no messengers are explicitly configured
- Make `LINE_CHANNEL_TOKEN` optional (only required when LINE is in enabled messengers)
- Wire terminal messenger handler into `cmd/server/main.go`
- Register terminal chat endpoint at `/api/chat/terminal`
- Update `.env.example` with new `ENABLED_MESSENGERS` variable

## Impact
- Affected specs: New `messenger-configuration` capability
- Affected code:
  - `internal/config/config.go` - Add `EnabledMessengers` field and validation logic
  - `cmd/server/main.go` - Conditional messenger initialization based on config
  - `.env.example` - Add `ENABLED_MESSENGERS` with terminal as default
