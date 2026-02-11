## ADDED Requirements

### Requirement: Intent Routing Before Expense Parsing
The system SHALL evaluate operational intents before attempting expense extraction, using locale-aware matching for supported command phrases.

#### Scenario: Route monthly report intent in Chinese
- **WHEN** user sends "本月報表"
- **THEN** system classifies the message as a report intent
- **AND** system executes report-link flow instead of expense parsing

#### Scenario: Route travel mode intent
- **WHEN** user sends "我正在日本旅行"
- **THEN** system classifies the message as travel-mode start intent
- **AND** system prepares currency context update flow rather than creating an expense

### Requirement: Pending Conversation State for Slot Filling
The system SHALL persist pending conversation state to support multi-turn clarification and slot completion.

#### Scenario: Currency clarification follow-up
- **WHEN** system asks "你要切換成哪個幣別？" for a currency-setting intent
- **AND** user replies "JPY"
- **THEN** system resolves the pending slot using the follow-up message
- **AND** system completes the original currency-setting intent

#### Scenario: Expired pending state
- **WHEN** pending clarification state has expired
- **AND** user sends a follow-up answer without restating intent
- **THEN** system does not apply the stale pending action
- **AND** system asks user to restate the request
