## 1. Configuration Layer
- [ ] 1.1 Add `EnabledMessengers` field to `Config` struct in `internal/config/config.go`
- [ ] 1.2 Add `ENABLED_MESSENGERS` env var parsing (comma-separated list, default: "terminal")
- [ ] 1.3 Update validation logic: only require LINE credentials when "line" is in enabled messengers
- [ ] 1.4 Add helper method `IsMessengerEnabled(name string) bool` to Config

## 2. Server Initialization
- [ ] 2.1 Wire terminal messenger handler in `cmd/server/main.go`
- [ ] 2.2 Register `/api/chat/terminal` endpoint for terminal messenger
- [ ] 2.3 Update LINE initialization to check `IsMessengerEnabled("line")` before requiring credentials
- [ ] 2.4 Update all other messengers to use `IsMessengerEnabled()` check pattern

## 3. Testing
- [ ] 3.1 Add unit tests for `EnabledMessengers` parsing in config
- [ ] 3.2 Add unit tests for `IsMessengerEnabled()` helper
- [ ] 3.3 Test server startup with `ENABLED_MESSENGERS=terminal` (no LINE credentials)
- [ ] 3.4 Test server startup with `ENABLED_MESSENGERS=line,telegram` (requires LINE credentials)

## 4. Documentation
- [ ] 4.1 Update `.env.example` with `ENABLED_MESSENGERS=terminal` as default
- [ ] 4.2 Update README.md with `ENABLED_MESSENGERS` environment variable documentation
