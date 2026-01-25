## ADDED Requirements
### Requirement: Unified Message Processing
The Messenger Gateway SHALL provide a unified interface for processing messages from various external platforms (LINE, Terminal, Discord, etc.), normalizing them into a standard internal format.

#### Scenario: Text Message Processing
- **WHEN** the Gateway receives a raw message payload from an external adapter
- **THEN** it identifies the source platform
- **AND** transforms the payload into a `UnifiedMessage` object
- **AND** routes the normalized message to the expense processing logic
- **AND** returns a standardized `MessageResponse`
