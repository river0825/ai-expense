# Draft: Edit Category Page

## Requirements (confirmed)
- Add a page to edit categories in the dashboard
- **Approach**: Integrate into Settings page with User Aggregate pattern (user decision)

## User Decision: User Aggregate Pattern
User wants categories as part of a unified User aggregate:
```
User
├── Categories (list of user's categories)
├── Home Currency  
└── Accounts (future)
```

## Research Findings

### Backend - User Aggregate Already Exists!
**`UserContext` aggregate** in `internal/domain/models.go`:
```go
type UserContext struct {
    User       *User
    Categories []*Category
}
```

This is already used for AI personalization. The aggregate pattern is in place!

### Backend Category API (Fully Implemented)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/categories` | List all categories for user |
| POST | `/api/categories` | Create new category |
| PUT | `/api/categories` | Update category |
| DELETE | `/api/categories` | Delete category |

**Domain Models** (`internal/domain/models.go`):
- `User`: UserID, MessengerType, HomeCurrency, Locale
- `Category`: ID, UserID, Name, IsDefault
- `CategoryKeyword`: Maps keywords to categories for AI suggestions
- `UserContext`: Aggregate containing User + Categories

### Frontend Current State
- **Settings page exists**: `/dashboard/settings` with currency preference
- **User model exists**: `domain/models/User.ts` with `User` and `UserSettings`
- **No category management UI** - only read-only aggregates for charts
- **RepositoryFactory pattern** in place

### Existing UI Patterns to Follow
- **Form handling**: `useState` (no react-hook-form)
- **API calls**: `RepositoryFactory` pattern
- **Styling**: Glass-panel with Tailwind (`glass-panel`, `bg-white/5`, `border-white/10`)
- **Icons**: `@heroicons/react/24/outline`
- **Settings page API**: `PUT /api/user/settings` via `HttpUserRepository.updateSettings`

## Technical Decisions
- **Location**: Add Categories section to existing Settings page (`/dashboard/settings`)
- **Pattern**: Extend Settings page (not new route) - aligns with User Aggregate concept
- **UI pattern**: Section with category list, inline edit, add/delete
- **State management**: `useState` for form fields
- **API integration**: Create `CategoryRepository` and add to `RepositoryFactory`

## Scope Boundaries
- **INCLUDE**:
  - Categories section in Settings page
  - List all user categories
  - Inline edit (rename category)
  - Add new category
  - Delete category (with confirmation, not default categories)
  - Update frontend User model to include categories
  
- **EXCLUDE**:
  - Category keyword management (AI suggestion keywords)
  - Category icon/color customization
  - Category reordering
  - Bulk operations
  - Accounts management (future feature)
  - Separate /categories route

## Resolved Questions
1. Inline vs modal edit? → **Inline** (matches ExpenseList pattern)
2. Keyword management? → **NO** (YAGNI)
3. Validation rules? → **Non-empty, max 50 chars**
4. Default categories editable? → **Rename YES, Delete NO**

## Test Strategy
- **Infrastructure exists**: YES (Vitest + Playwright)
- **Approach**: Tests-after (not TDD, but unit tests for repository)
- **Unit tests**: Vitest for CategoryRepository logic
- **E2E tests**: Playwright for category management flow
- **Agent QA**: Required for all tasks (Playwright for UI verification)
