## MODIFIED Requirements
### Requirement: AI Service Cost Management
The system SHALL track and persist token usage and estimated costs for every AI API interaction to enable auditing and budget management.

#### Scenario: Persist cost for successful request
- **WHEN** AI service successfully parses a message
- **THEN** system calculates cost based on input/output tokens and model pricing
- **AND** persists a cost log entry with user_id, operation_type, and cost
- **AND** returns the result to the caller

#### Scenario: Persist cost for failed request
- **WHEN** AI service receives a response but fails to parse content (e.g. empty JSON)
- **THEN** system still records the token usage and cost
- **BECAUSE** the API provider still charges for the tokens used

#### Scenario: Cache parsed results
- **WHEN** same text is parsed multiple times
- **THEN** system returns cached result instead of calling AI again
- **AND** no new cost log is created for cache hits
- **AND** cache expires after 24 hours

#### Scenario: Batch processing for efficiency
- **WHEN** parsing multiple expenses in one message
- **THEN** system uses single API call if possible
- **AND** extracts multiple items from one response
