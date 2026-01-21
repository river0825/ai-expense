# Change: Add Privacy Policy & Terms of Use Endpoint

## Why
To comply with legal requirements and transparency standards, the application needs to serve Privacy Policy and Terms of Use documents. These documents should be retrievable from the database to allow easier updates without redeployment.

## What Changes
- Add `Policy` domain model to store legal document content.
- Add `PolicyRepository` to retrieve policies by key (e.g., "privacy_policy", "terms_of_use").
- Create a new database table `policies` to store the content.
- Expose a public API endpoint `GET /api/policies/{key}` to serve these documents.
- Seed initial content for Privacy Policy and Terms of Use.

## Impact
- **Specs**: `legal-compliance` (Added)
- **Code**:
  - `internal/domain`: New model and repository.
  - `internal/adapter/repository`: New SQLite implementation.
  - `internal/adapter/http`: New handler for policies.
  - `migrations`: New table.
