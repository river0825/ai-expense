## ADDED Requirements

### Requirement: Short Link Generation
The system SHALL generate temporary short links for secure report access.

#### Scenario: Generate Short Link for Report
Given a user requests to view their expense report
When the system processes the request
Then it generates a signed JWT containing the user ID and expiration
And it generates a unique, random 6-character Short ID
And it stores the mapping of Short ID to JWT with a 5-minute expiration
And it returns a URL in the format `https://<domain>/r/<short_id>`

### Requirement: Short Link Redirection
The system SHALL validate short links and redirect users to the dashboard.

#### Scenario: Redirect Short Link
Given a user accesses a valid short link URL
When the system validates the Short ID
And the link has not expired
Then it redirects the user to the Dashboard Report page
And it sets the JWT as a secure, HTTP-only cookie
Or it includes the JWT as a query parameter for the frontend to handle

#### Scenario: Expired Short Link
Given a user accesses an expired or invalid short link
When the system attempts to validate the ID
Then it returns a 404 Not Found or a specific error page indicating expiration
