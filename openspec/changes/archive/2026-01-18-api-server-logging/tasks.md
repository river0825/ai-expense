## 1. Middleware Implementation
- [x] 1.1 Create `internal/adapter/http/logging.go`
- [x] 1.2 Implement `LoggingMiddleware` that wraps `http.Handler`
- [x] 1.3 Capture status code using a custom `ResponseWriter` wrapper
- [x] 1.4 Log details: Time, Method, URI, Status, Duration, UserAgent

## 2. Server Integration
- [x] 2.1 Update `cmd/server/main.go` to wrap the router with `LoggingMiddleware`
- [x] 2.2 Ensure it runs *before* CORS middleware (or after, depending on preference for logging blocked requests) - usually outermost.

## 3. Verification
- [x] 3.1 Verify logs appear in stdout when making requests
