## 1. Configuration Layer
- [x] 1.1 Add `EnabledMessengers` field to `Config` struct in `internal/config/config.go`
- [x] 1.2 Add `ENABLED_MESSENGERS` env var parsing (comma-separated list, default: "terminal")
- [x] 1.3 Update validation logic: only require LINE credentials when "line" is in enabled messengers
- [x] 1.4 Add helper method `IsMessengerEnabled(name string) bool` to Config

## 2. Server Initialization
- [x] 2.1 Wire terminal messenger handler in `cmd/server/main.go`
- [x] 2.2 Register `/api/chat/terminal` endpoint for terminal messenger
- [x] 2.3 Update LINE initialization to check `IsMessengerEnabled("line")` before requiring credentials
- [x] 2.4 Update all other messengers to use `IsMessengerEnabled()` check pattern

## 3. Testing
- [x] 3.1 Add unit tests for `EnabledMessengers` parsing in config
- [x] 3.2 Add unit tests for `IsMessengerEnabled()` helper
- [x] 3.3 Test server startup with `ENABLED_MESSENGERS=terminal` (no LINE credentials)
- [x] 3.4 Test server startup with `ENABLED_MESSENGERS=line,telegram` (requires LINE credentials)

## 4. Documentation
- [x] 4.1 Update `.env.example` with `ENABLED_MESSENGERS=terminal` as default
- [x] 4.2 Update README.md with `ENABLED_MESSENGERS` environment variable documentation
