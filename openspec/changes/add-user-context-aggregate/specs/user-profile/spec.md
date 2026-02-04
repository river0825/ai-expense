## ADDED Requirements
### Requirement: User Context Aggregate
The system SHALL provide a user context aggregate that includes user settings and the user's category list for prompt personalization.

#### Scenario: Load user context by user_id
- **WHEN** the system requests user context for a user_id
- **THEN** it returns settings (home currency, locale) and the full list of the user's categories

### Requirement: Prompt Builder Access to User Context
The system SHALL allow AI prompt builders to consume user context when constructing parsing prompts.

#### Scenario: Personalized prompt inputs
- **WHEN** a parsing prompt is constructed
- **THEN** the prompt builder has access to user home currency and category names
