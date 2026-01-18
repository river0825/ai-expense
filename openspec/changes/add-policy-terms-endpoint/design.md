## Context
The system requires a way to serve legal documents like Privacy Policy and Terms of Use. These documents change over time, so versioning and database storage are preferred over hardcoding.

## Decisions
- **Decision**: Store policies in a `policies` table with a unique `key` (e.g., 'privacy_policy', 'terms_of_use').
- **Why**: Allows flexible addition of other document types later.
- **Decision**: Use a simple REST endpoint `GET /api/policies/{key}`.
- **Why**: Standard pattern, easy for frontend to consume.
- **Decision**: Include `version` column.
- **Why**: To track changes and potentially show users which version they agreed to (future scope).

## Data Model
### Policy
- `id`: UUID (PK)
- `key`: String (Unique, e.g., "privacy_policy")
- `title`: String
- `content`: Text (Markdown or HTML)
- `version`: String (e.g., "1.0")
- `created_at`: Timestamp
- `updated_at`: Timestamp

## API Interface
### GET /api/policies/{key}
**Response:**
```json
{
  "key": "privacy_policy",
  "title": "Privacy Policy",
  "content": "...",
  "version": "1.0",
  "last_updated": "2024-01-20T10:00:00Z"
}
```
