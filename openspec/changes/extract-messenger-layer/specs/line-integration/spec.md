## MODIFIED Requirements
### Requirement: Receive Messages from LINE Messaging API
The system SHALL receive incoming messages from LINE users through the Messaging API webhook and delegate processing to the Messenger Gateway.

#### Scenario: Webhook receives and delegates message
- **WHEN** user sends text message in LINE bot chat
- **THEN** system receives webhook event
- **AND** system validates the signature
- **AND** system delegates the event to the Messenger Gateway

#### Scenario: Support various message types
- **WHEN** user sends text, but also possibly sends stickers or locations
- **THEN** system prioritizes text messages and gracefully ignores unsupported types before delegation

#### Scenario: Webhook responds within timeout
- **WHEN** LINE sends webhook to bot endpoint
- **THEN** system responds with HTTP 200 within 3 seconds
