## 1. Middleware Implementation
- [ ] 1.1 Create `internal/adapter/http/logging.go`
- [ ] 1.2 Implement `LoggingMiddleware` that wraps `http.Handler`
- [ ] 1.3 Capture status code using a custom `ResponseWriter` wrapper
- [ ] 1.4 Log details: Time, Method, URI, Status, Duration, UserAgent

## 2. Server Integration
- [ ] 2.1 Update `cmd/server/main.go` to wrap the router with `LoggingMiddleware`
- [ ] 2.2 Ensure it runs *before* CORS middleware (or after, depending on preference for logging blocked requests) - usually outermost.

## 3. Verification
- [ ] 3.1 Verify logs appear in stdout when making requests
