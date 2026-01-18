## ADDED Requirements
### Requirement: HTTP Request Logging
The system SHALL log metadata for every incoming HTTP request to stdout.

#### Scenario: Successful Request
- **WHEN** a client makes a valid API request (e.g., `GET /health`)
- **THEN** the server logs the method, path, status code (200), and latency
- **AND** the log entry includes the timestamp

#### Scenario: Failed Request
- **WHEN** a client makes an invalid request (e.g., `GET /unknown`)
- **THEN** the server logs the method, path, status code (404), and latency

#### Scenario: Server Error
- **WHEN** an internal error occurs (500)
- **THEN** the server logs the status code 500 and the request details
