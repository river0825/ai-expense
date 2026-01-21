# ai-service Specification

## Purpose
TBD - created by archiving change add-conversational-expense-tracker. Update Purpose after archive.
## Requirements
### Requirement: AI-Powered Conversation Parsing
The system SHALL use an AI model to intelligently parse natural language expense input, supporting complex formats and implicit context.

#### Scenario: Parse multi-format expenses
- **WHEN** user sends "早餐20，午餐30，加油200" (no $ symbols)
- **THEN** system uses Gemini 2.5 lite to understand and extract: (早餐, 20), (午餐, 30), (加油, 200)

#### Scenario: Handle implicit amounts
- **WHEN** user sends "剛才買咖啡，花了不少" (vague amount)
- **THEN** system asks for clarification: "咖啡花了多少錢?"

#### Scenario: Understand context from date references
- **WHEN** user sends "上禮拜去度假，住宿5000、餐費2000"
- **THEN** system extracts date=last week, expenses: (住宿, 5000), (餐費, 2000)

#### Scenario: Fallback to regex parsing
- **WHEN** AI service is unavailable or rate-limited
- **THEN** system falls back to regex-based parsing
- **AND** continues to process user requests

### Requirement: Pluggable AI Implementation
The system SHALL provide an interface-based AI service that allows swapping between different AI models without changing business logic.

#### Scenario: Current implementation uses Gemini 2.5 lite
- **WHEN** system initializes
- **THEN** AIService is implemented by GeminiAI client
- **AND** uses Google Generative AI API with model 'gemini-2.5-lite'

#### Scenario: Swap to different AI provider
- **WHEN** operator changes configuration (e.g., CLAUDE_API_KEY is set)
- **THEN** system switches to ClaudeAI implementation
- **AND** no code changes required (only configuration)
- **AND** existing expenses and parsing behavior unchanged

#### Scenario: Future model options
- **WHEN** new AI models become available (e.g., GPT-5, Gemini 3.0)
- **THEN** new implementations (GPT5AI, Gemini3AI) can be added
- **AND** swapped via configuration without modifying use cases

#### Scenario: Local LLM fallback
- **WHEN** external AI providers are unavailable
- **THEN** system can use local LLM (e.g., Ollama, LlamaCPP)
- **AND** LocalLLM implementation provides reasonable accuracy

### Requirement: AI-Powered Category Suggestion
The system SHALL use AI to intelligently suggest expense categories based on description, improving upon keyword-based matching.

#### Scenario: Smart category inference
- **WHEN** user enters expense "買Starbucks咖啡"
- **THEN** Gemini suggests "Food" (understands coffee is food-related)
- **AND** user can accept or override

#### Scenario: Handle ambiguous descriptions
- **WHEN** user enters "Apple" (could be fruit or tech company)
- **THEN** system asks clarifying question: "買Apple - 是水果還是產品?"
- **OR** suggests most likely category based on context

#### Scenario: Learn from user overrides
- **WHEN** user frequently overrides suggestions for certain keywords
- **THEN** system can learn (optional: add to keyword mapping)

#### Scenario: Multilingual category understanding
- **WHEN** user enters expense in Chinese or English mixed text
- **THEN** Gemini handles code-switching naturally

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

### Requirement: AI Service Configuration & Flexibility
The system SHALL allow configuration of which AI provider and model to use without code changes.

#### Scenario: Configuration via environment variables
- **WHEN** environment sets `AI_PROVIDER=gemini` and `AI_MODEL=gemini-2.5-lite`
- **THEN** system uses those settings on startup
- **WHEN** environment sets `AI_PROVIDER=claude` and `CLAUDE_API_KEY=...`
- **THEN** system switches to Claude implementation

#### Scenario: API key management
- **WHEN** AI service needs API credentials
- **THEN** system reads from environment variables or secure config
- **AND** never hardcodes keys

#### Scenario: Fallback configuration
- **WHEN** AI provider is unavailable
- **THEN** system uses configured fallback (e.g., regex parsing)
- **AND** continues operation in degraded mode

### Requirement: AI Service Error Handling
The system SHALL handle AI service failures gracefully with fallback strategies.

#### Scenario: Rate limit handling
- **WHEN** AI API hits rate limit
- **THEN** system falls back to regex parsing
- **AND** queues request for retry

#### Scenario: API timeout
- **WHEN** AI request times out after 5 seconds
- **THEN** system uses faster regex parsing
- **AND** returns approximate result

#### Scenario: User communication on AI failure
- **WHEN** AI service fails
- **THEN** user doesn't see error details
- **AND** response is still generated (using fallback)
- **AND** message like "處理中，請稍候" is avoided

