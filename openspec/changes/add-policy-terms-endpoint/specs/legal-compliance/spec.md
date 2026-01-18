## ADDED Requirements
### Requirement: Legal Document Management
The system SHALL store and serve legal documents such as Privacy Policy and Terms of Use from the database.

#### Scenario: Retrieve Privacy Policy
- **WHEN** user requests "Privacy Policy" (key: privacy_policy)
- **THEN** system returns the latest content of the Privacy Policy
- **AND** includes version and last updated timestamp

#### Scenario: Retrieve Terms of Use
- **WHEN** user requests "Terms of Use" (key: terms_of_use)
- **THEN** system returns the latest content of the Terms of Use

#### Scenario: Document Not Found
- **WHEN** user requests a non-existent policy key
- **THEN** system returns a 404 Not Found error
