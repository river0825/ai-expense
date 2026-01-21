## ADDED Requirements

### Requirement: Webchat Interface
The dashboard SHALL provide a chat interface to interact with the expense tracking bot.

#### Scenario: Send message
- **WHEN** user types a message and clicks send
- **THEN** message is displayed in history
- **AND** bot response is fetched and displayed

#### Scenario: Session persistence
- **WHEN** user reloads the page
- **THEN** the same user_id is maintained
