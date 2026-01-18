# messenger-configuration Specification

## Purpose
TBD - created by archiving change add-messenger-env-toggle. Update Purpose after archive.
## Requirements
### Requirement: Messenger Selection via Environment Variable
The system SHALL allow operators to specify which messengers are enabled via the `ENABLED_MESSENGERS` environment variable. The `.env.example` file SHALL document this variable with `terminal` as the default value.

#### Scenario: Default to terminal messenger
- **WHEN** the server starts without `ENABLED_MESSENGERS` set
- **THEN** the terminal messenger SHALL be enabled by default
- **AND** the server SHALL start successfully without LINE credentials
- **AND** developers can immediately test expense tracking via terminal

#### Scenario: Enable specific messengers
- **WHEN** `ENABLED_MESSENGERS` is set to "line,telegram"
- **THEN** only LINE and Telegram messengers SHALL be initialized
- **AND** other messengers (discord, slack, teams, whatsapp, terminal) SHALL NOT be initialized

#### Scenario: Enable terminal for local development
- **WHEN** `ENABLED_MESSENGERS` is set to "terminal"
- **THEN** the terminal messenger SHALL be enabled
- **AND** the `/api/chat/terminal` endpoint SHALL be available
- **AND** no external messenger credentials SHALL be required

### Requirement: Conditional Credential Validation
The system SHALL only require messenger-specific credentials when that messenger is enabled.

#### Scenario: LINE credentials required only when LINE enabled
- **WHEN** `ENABLED_MESSENGERS` includes "line"
- **THEN** `LINE_CHANNEL_TOKEN` SHALL be required
- **AND** the server SHALL fail to start if `LINE_CHANNEL_TOKEN` is missing

#### Scenario: LINE credentials not required when LINE disabled
- **WHEN** `ENABLED_MESSENGERS` does not include "line"
- **THEN** `LINE_CHANNEL_TOKEN` SHALL NOT be required
- **AND** the server SHALL start successfully without LINE credentials

#### Scenario: Multiple messengers with partial credentials
- **WHEN** `ENABLED_MESSENGERS` is set to "terminal,telegram"
- **AND** `TELEGRAM_BOT_TOKEN` is not set
- **THEN** the server SHALL log a warning for Telegram
- **AND** terminal messenger SHALL still be enabled
- **AND** the server SHALL start successfully

### Requirement: Terminal Messenger Endpoint Registration
The system SHALL register the terminal messenger HTTP endpoint when terminal is enabled.

#### Scenario: Terminal chat endpoint available
- **WHEN** terminal messenger is enabled
- **THEN** `POST /api/chat/terminal` SHALL accept message requests
- **AND** `GET /api/chat/terminal/user` SHALL return user information

#### Scenario: Terminal chat endpoint unavailable when disabled
- **WHEN** terminal messenger is not enabled
- **THEN** `POST /api/chat/terminal` SHALL return 404 Not Found

