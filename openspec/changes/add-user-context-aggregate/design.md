## Context
User settings (home currency, locale) live on the user record while categories are separate entities. AI prompts are currently static and cannot easily consume user-specific context.

## Goals / Non-Goals
- Goals: Provide a single user context aggregate that includes settings and categories; enable prompt builders to consume it.
- Non-Goals: Introduce new settings fields or change existing database schema.

## Decisions
- Decision: Create a domain-level user context struct that embeds core user settings and a category list.
- Alternatives considered: Embedding categories directly into User repository queries (rejected to avoid cross-repo coupling).

## Risks / Trade-offs
- Additional queries to assemble user context (user + categories) → mitigate with simple caching later if needed.

## Migration Plan
- No schema changes required; load categories and settings from existing tables.
- Update AI prompt generation to optionally use the user context aggregate.

## Open Questions
- Should notification preferences be included in the aggregate now or later?
