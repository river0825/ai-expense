## ADDED Requirements

### Requirement: Independent Admin Authentication
The system SHALL provide authentication for admin analytics that is independent from end-user report access flows.

#### Scenario: Admin login succeeds with valid credentials
- **WHEN** admin submits valid login credentials to admin auth endpoint
- **THEN** system creates authenticated admin session/token

#### Scenario: Admin login fails with invalid credentials
- **WHEN** admin submits invalid credentials
- **THEN** system returns `401 Unauthorized` without exposing credential validation internals

#### Scenario: End-user report token is rejected for admin analytics
- **WHEN** client attempts admin analytics access using end-user report token/cookie
- **THEN** system denies access with `401 Unauthorized`

### Requirement: Authorization for Analytics Access
The system SHALL enforce authorization for protected analytics resources after successful authentication.

#### Scenario: Authorized admin accesses analytics endpoints
- **WHEN** authenticated admin with analytics permission requests protected endpoint
- **THEN** system returns requested analytics payload

#### Scenario: Authenticated but unauthorized principal is blocked
- **WHEN** authenticated principal without required permission requests protected endpoint
- **THEN** system returns `403 Forbidden`

### Requirement: Session Lifecycle and Security Controls
The system SHALL enforce secure admin session lifecycle behavior.

#### Scenario: Session expiration
- **WHEN** admin session/token expires
- **THEN** subsequent protected requests return `401 Unauthorized`

#### Scenario: Explicit logout
- **WHEN** admin logs out
- **THEN** session/token is invalidated and cannot access protected analytics routes

#### Scenario: Auditability for auth events
- **WHEN** admin login, logout, and protected access attempts occur
- **THEN** system records security-relevant audit events without storing sensitive credentials
