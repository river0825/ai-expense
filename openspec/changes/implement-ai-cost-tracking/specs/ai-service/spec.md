## MODIFIED Requirements

### Requirement: AI Service Cost Management
The system SHALL track and persist actual token usage and calculated costs for every AI API interaction, with pricing stored in the database and cost calculation performed in the business logic layer.

#### Scenario: Persist actual token counts for successful request
- **WHEN** AI service successfully parses a message and returns token metadata
- **THEN** system extracts actual input and output token counts from API response
- **AND** persists a cost log entry with real token counts, user_id, operation_type, and cost
- **AND** returns the result to the caller

#### Scenario: Calculate cost from token usage and pricing lookup
- **WHEN** cost log entry is created with actual token counts
- **THEN** system looks up provider/model pricing from database
- **AND** calculates cost as: (input_tokens * input_price + output_tokens * output_price) / 1,000,000
- **AND** stores calculated cost in USD with currency field

#### Scenario: Handle missing or inactive pricing configuration
- **WHEN** pricing is not found for a provider/model combination
- **THEN** system logs a warning
- **AND** saves cost as 0 with cost_note field set to "pricing_not_configured"
- **AND** continues processing without blocking the user

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
- **AND** logs single cost entry for the batch (tokens are for entire call)

## ADDED Requirements

### Requirement: AI Service Token Metadata in Responses
The system SHALL return actual token usage information from AI API responses, enabling cost calculation in the business logic layer.

#### Scenario: Token metadata included in parse response
- **WHEN** AI service successfully calls Gemini API and receives response
- **THEN** response includes token metadata: input_tokens, output_tokens, total_tokens
- **AND** tokens come from actual API response metadata, not estimates
- **AND** usecase layer can calculate cost from this metadata

#### Scenario: Token metadata for category suggestion
- **WHEN** AI service suggests category using Gemini API
- **THEN** response includes token metadata (or zero tokens if keyword fallback used)
- **AND** usecase layer logs cost for API calls only

#### Scenario: Fallback parsing includes zero token metadata
- **WHEN** AI API is unavailable and system falls back to regex parsing
- **THEN** response includes zero token metadata
- **AND** cost log shows 0 tokens, 0 cost (no API call made)

### Requirement: AI Provider Pricing Configuration
The system SHALL support flexible pricing configuration for different AI providers and models, enabling accurate cost tracking without code changes.

#### Scenario: Price lookup by provider and model
- **WHEN** cost calculation is performed for Gemini 2.5 lite
- **THEN** system queries pricing table with provider="gemini" and model="gemini-2.5-lite"
- **AND** retrieves input_token_price and output_token_price (per 1M tokens)
- **AND** uses active pricing where is_active=true and effective_date <= current UTC timestamp

#### Scenario: Support different input/output token pricing
- **WHEN** provider charges differently for input vs output tokens (as Gemini does)
- **THEN** pricing table stores separate prices: input_token_price and output_token_price
- **AND** calculation uses both: (input * in_price + output * out_price) / 1,000,000

#### Scenario: USD currency for all costs
- **WHEN** cost is calculated and stored
- **THEN** currency field is always "USD"
- **AND** provider prices are stored in USD per 1M tokens

#### Scenario: Add new provider without code changes
- **WHEN** operator wants to use a new AI provider (Claude, OpenAI, etc.)
- **THEN** admin inserts new rows in pricing table with provider/model/pricing
- **AND** system automatically supports the new provider in cost calculations
- **AND** no code deployment required

#### Scenario: Inactive pricing configuration
- **WHEN** pricing becomes outdated (e.g., provider drops a model)
- **THEN** is_active flag is set to false in database
- **AND** queries ignore inactive rows
- **AND** requests for inactive models log warning and cost as 0

### Requirement: AI Cost Log Note Field
The system SHALL store optional notes in cost logs to document special conditions (e.g., missing pricing configuration).

#### Scenario: Cost note for missing pricing
- **WHEN** pricing configuration is not found during cost calculation
- **THEN** cost_note field is populated with "pricing_not_configured"
- **AND** cost is saved as 0
- **AND** entry remains auditable and traceable

#### Scenario: Cost note for other conditions
- **WHEN** other special conditions occur during cost logging
- **THEN** cost_note field can store descriptive text (e.g., "fallback_parsing", "cache_hit")
- **AND** NULL cost_note indicates normal operation
