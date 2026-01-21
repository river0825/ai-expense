## Context
We need visibility into API traffic. Standard Go `net/http` doesn't log requests by default.

## Decisions
- **Decision**: Use standard library `log` with structured text format for now.
- **Why**: Keeps dependencies low. Can upgrade to `slog` or `zap` later if needed.
- **Format**: `[API] timestamp | status | duration | method path`
- **Output**: Stdout (Docker friendly).

## Middleware Design
Wrapper around `http.ResponseWriter` to capture the status code, as it's not exposed by default after writing.

```go
type responseWriter struct {
    http.ResponseWriter
    status int
    size   int
}
```
