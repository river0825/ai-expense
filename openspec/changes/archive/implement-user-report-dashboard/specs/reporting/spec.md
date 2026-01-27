## MODIFIED Requirements

### Requirement: Dynamic Report Generation
The system SHALL generate expense reports based on dynamic date ranges and user context.

#### Scenario: Fetch Report Summary with Date Range
Given an authenticated user on the report dashboard
When the user selects a specific date range (e.g., "Last 7 Days", "This Month")
And the frontend requests the report summary with `start_date` and `end_date` parameters
Then the system returns the total expense for that period
And it returns a list of expenses within that period
And it returns expense aggregation by category for that period
And it excludes "Total Income" and "Total Balance" metrics (as requested)

### Requirement: Secure API Access
The API SHALL strictly validate access tokens for report data.

#### Scenario: Secure API Access
Given an API request to `/api/reports/summary`
When the request contains a valid JWT in the Cookie or Authorization header
Then the system validates the token signature and expiration
And it extracts the `user_id` from the token claims
And it fetches data ONLY for that user ID
