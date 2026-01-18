# Change: API Server Logging

## Why
The API server currently uses minimal logging (`log.Printf` only for startup and major errors). To properly debug issues, monitor traffic, and audit access, we need structured logging that captures request details (method, path, status, duration) and errors.

## What Changes
- Add a logging middleware to intercept all HTTP requests.
- Log request method, path, remote address, status code, and duration.
- Use structured JSON logging for easier parsing by monitoring tools (optional but recommended for production).
- Update the main server setup to use this middleware.

## Impact
- **Specs**: `server-operations` (Added)
- **Code**:
  - `internal/adapter/http`: New `logging.go` middleware.
  - `cmd/server/main.go`: Apply middleware.
