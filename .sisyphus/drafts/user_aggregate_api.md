# Draft: User Aggregate Settings API

## Requirements (confirmed)
- [New API]: Get user aggregate settings (accounts, home currency, categories, etc.) in one call.
- [New API]: Save user aggregate settings in one call.
- [Frontend]: Use new APIs to show/edit settings.
- [Constraint]: Expense API remains unchanged.
- [Infrastructure]: Use `git worktree` for isolation (User explicitly requested).
- [Save Logic]: Full Replacement (Client sends full state, server replaces lists).

## Research Findings
- [Pending] Researching backend models for User, Account, Category.
- [Pending] Researching frontend components for settings.

## Research Findings
- [Pending] Researching backend models for User, Account, Category.
- [Pending] Researching frontend components for settings.

## Open Questions
- What exact fields are in "user aggregate"? (Need to find current models)
- Does "save" mean replacing the whole aggregate or partial updates?
- "create a new worktree" -> Do you mean `git worktree` for isolation, or just a new branch?
