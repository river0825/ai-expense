# Group Expense Splitting Design

> Created: 2026-01-30
> Status: Approved

## Overview

Add group expense splitting functionality to AIExpense. When the chatbot is added to a Messenger group, it enables expense tracking for the group. Members can input expenses, and the system calculates who owes whom.

## Design Decisions

| Item | Decision |
|------|----------|
| Use Cases | Universal (travel, roommates, dining - all scenarios) |
| Group Creation | Hybrid: Auto-detect from Messenger + Dashboard management |
| Split Method | Default equal split, customizable per-expense in Dashboard |
| Member Scope | Per-expense setting, defaults to all group members |
| Settlement | Record actual consumption to personal ledgers + archive group ledger |
| Post-Archive Missed Expenses | Record in new ledger (archive is irreversible) |
| Disable Splitting | Bot goes silent, but first shows summary + Dashboard link |
| Re-enable Splitting | User chooses: continue previous ledger or start new |
| Permissions | All members can edit, full audit log of all changes |
| Settlement Prerequisite | All splitting members must be registered |
| Transfer Suggestions | Show direct debt relationships (no optimization) |
| i18n | Full internationalization support |

## Domain Models

```go
// Group represents a group ledger
type Group struct {
    ID            string
    Name          string
    SourceType    string    // "line", "discord", "telegram", "manual"
    SourceID      string    // Messenger group ID (for auto-detection)
    Status        string    // "active", "closed", "archived"
    CreatedAt     time.Time
    CreatedBy     string    // UserID
}

// GroupMember represents a group member
type GroupMember struct {
    GroupID   string
    UserID    string
    JoinedAt  time.Time
}

// GroupExpense represents a group expense
type GroupExpense struct {
    ID           string
    GroupID      string
    PaidByUserID string    // Who paid
    Description  string
    Amount       float64
    ExpenseDate  time.Time
    CreatedAt    time.Time
    CreatedBy    string    // Who created this record (may differ from payer)
}

// GroupExpenseSplit represents how an expense is split among members
type GroupExpenseSplit struct {
    GroupExpenseID string
    UserID         string
    Amount         float64   // Amount this member should pay
}

// GroupSettlement represents a settlement record
type GroupSettlement struct {
    ID          string
    GroupID     string
    SettledAt   time.Time
    SettledBy   string    // Who performed the settlement
}

// GroupSettlementDebt represents debt details at settlement time
type GroupSettlementDebt struct {
    SettlementID string
    FromUserID   string    // Who pays
    ToUserID     string    // Pays to whom
    Amount       float64
}

// GroupAuditLog represents an audit log entry
type GroupAuditLog struct {
    ID         string
    GroupID    string
    UserID     string    // Who performed the action
    Action     string    // "create_expense", "update_expense", "delete_expense", "settle", etc.
    TargetType string    // "expense", "member", "group"
    TargetID   string
    OldValue   string    // JSON
    NewValue   string    // JSON
    CreatedAt  time.Time
}
```

## Chatbot Interaction Flow

### When Bot Joins Group

```
Trigger: Bot is added to a Messenger group
Response:
  "Expense splitting is now active!
   
   Simply type expenses to record them, e.g.:
   "Lunch $350" "Taxi 200"
   
   Type "disable expense splitting" to turn off
   
   Group ledger: {dashboard_url}"
```

### Recording Expenses

```
Trigger: Group member types "Dinner $900"
Process:
  1. Identify sender → set as PaidByUserID
  2. AI parses amount and description
  3. Create GroupExpense (default: all members split equally)
  4. Create GroupExpenseSplit (each member's share)
  
Response:
  "Recorded!
   
   Dinner $900 (Ming paid)
   
   Current split status:
   ┌──────────┬─────────┬─────────┬─────────┐
   │ Member   │ Paid    │ Owes    │ Receives│
   ├──────────┼─────────┼─────────┼─────────┤
   │ Ming     │ $900    │ $0      │ $600    │
   │ Hua      │ $0      │ $300    │ $0      │
   │ Mei      │ $0      │ $300    │ $0      │
   └──────────┴─────────┴─────────┴─────────┘
   
   Details: {dashboard_url}"
```

### Disabling Expense Splitting

```
Trigger: Any member types "disable expense splitting"
Process:
  1. Set Group.Status to "closed"
  2. Calculate current debts

Response:
  "Expense splitting disabled
   
   Current settlement:
   • Hua → Ming $100
   • Mei → Ming $300  
   • Mei → Hua $200
   
   Please complete settlement at the group ledger:
   {dashboard_url}
   
   Type "enable expense splitting" to start a new ledger"
```

### Re-enabling Expense Splitting

```
Trigger: Any member types "enable expense splitting"

Case A - Has unsettled ledger:
  "Found unsettled ledger (3 expenses, total $1,500)
   
   Please choose:
   1. Type "continue" to resume previous ledger
   2. Type "new ledger" to start fresh (old ledger will be archived)"

Case B - No previous ledger:
  "Expense splitting is now active!
   
   Simply type expenses to record them
   Group ledger: {dashboard_url}"
```

## Group Dashboard Features

### Page Structure

```
/[locale]/groups/[groupId]
├── Summary Section
├── Expense List
├── Member Management
├── Settlement Section
└── Audit Log
```

### Summary Section

Displays overall group ledger status:

```
┌─────────────────────────────────────────────────────┐
│  Japan Trip                            [Edit Name]  │
│  Status: Active                                     │
│                                                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │
│  │ Total       │ │ Expenses    │ │ Members     │   │
│  │ $1,500      │ │ 2           │ │ 3           │   │
│  └─────────────┘ └─────────────┘ └─────────────┘   │
│                                                     │
│  Split Status:                                      │
│  ┌──────────┬─────────┬─────────┬─────────┐        │
│  │ Member   │ Paid    │ Owes    │ Receives│        │
│  ├──────────┼─────────┼─────────┼─────────┤        │
│  │ Ming     │ $900    │ $0      │ $400    │        │
│  │ Hua      │ $600    │ $0      │ $100    │        │
│  │ Mei      │ $0      │ $500    │ $0      │        │
│  └──────────┴─────────┴─────────┴─────────┘        │
└─────────────────────────────────────────────────────┘
```

### Expense List

Expandable to view and edit split details:

```
┌─────────────────────────────────────────────────────┐
│  Expenses                             [+ Add New]   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ Dinner              $900    Ming    01/15   │   │
│  │    Split: Ming, Hua, Mei ($300 each)  [Edit]│   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ Tickets             $600    Hua     01/15   │   │
│  │    Split: Ming, Hua, Mei ($200 each)  [Edit]│   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

**Editable fields:**
- Description, amount, date
- Payer
- Split members (select who shares this expense)
- Per-member split amounts (default equal, customizable)

### Member Management

Manage group ledger participants:

```
┌─────────────────────────────────────────────────────┐
│  Members                              [+ Add]       │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ Ming    LINE: @ming                 [Remove]│   │
│  │ Hua     LINE: @hua                  [Remove]│   │
│  │ Mei     LINE: @mei                  [Remove]│   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  New expenses default to splitting among all above  │
└─────────────────────────────────────────────────────┘
```

### Settlement Section

Transfer suggestions and settle button:

```
┌─────────────────────────────────────────────────────┐
│  Settlement                                         │
│                                                     │
│  Transfer Suggestions:                              │
│  ┌─────────────────────────────────────────────┐   │
│  │ Hua → Ming    $100                          │   │
│  │ Mei → Ming    $300                          │   │
│  │ Mei → Hua     $200                          │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  After settlement:                                  │
│  • Each expense will be recorded to members'        │
│    personal ledgers based on split ratio            │
│  • This group ledger will be archived (read-only)   │
│                                                     │
│                              [Settle and Archive]   │
└─────────────────────────────────────────────────────┘
```

### Audit Log

Shows all changes:

```
┌─────────────────────────────────────────────────────┐
│  Audit Log                                          │
│                                                     │
│  01/15 14:32  Ming added expense "Dinner $900"      │
│  01/15 15:10  Hua added expense "Tickets $600"      │
│  01/15 16:45  Mei updated "Dinner" amount $900→$950 │
│  01/15 16:45  Mei updated "Dinner" split: removed   │
│               Hua                                   │
│                                                     │
│                                      [View More...] │
└─────────────────────────────────────────────────────┘
```

## API Design

### Group Management

```
GET    /api/groups                    # List user's groups
GET    /api/groups/:groupId           # Get group details
PUT    /api/groups/:groupId           # Update group (name, etc.)
POST   /api/groups/:groupId/merge     # Merge groups (advanced)
```

### Member Management

```
POST   /api/groups/:groupId/members            # Add member
DELETE /api/groups/:groupId/members/:userId    # Remove member
```

### Expense Management

```
POST   /api/groups/:groupId/expenses              # Add expense
PUT    /api/groups/:groupId/expenses/:expenseId   # Update expense
DELETE /api/groups/:groupId/expenses/:expenseId   # Delete expense
```

### Settlement

```
GET    /api/groups/:groupId/settlement/preview    # Preview settlement
POST   /api/groups/:groupId/settle                # Execute settlement
```

### Audit Log

```
GET    /api/groups/:groupId/audit-logs    # Get audit logs
```

### Webhook Extension (Messenger Integration)

Existing Messenger webhook handlers need extension:

```
1. Check if message source is a group
2. If group:
   - Check if Group record exists for this messenger group
   - If not → auto-create Group (Status: active)
   - Check Group.Status
     - active → parse and record expense
     - closed → no response
3. If personal message → maintain existing logic (record to personal ledger)
```

## File Structure

```
internal/domain/
├── group.go                    # Group domain models

internal/usecase/group/
├── create_group.go
├── create_group_test.go
├── get_group.go
├── get_group_test.go
├── update_group.go
├── update_group_test.go
├── add_member.go
├── add_member_test.go
├── remove_member.go
├── remove_member_test.go
├── create_expense.go
├── create_expense_test.go
├── update_expense.go
├── update_expense_test.go
├── delete_expense.go
├── delete_expense_test.go
├── calculate_balances.go
├── calculate_balances_test.go
├── calculate_debts.go
├── calculate_debts_test.go
├── settle.go
├── settle_test.go
├── log_audit.go
├── log_audit_test.go
├── process_message.go
└── process_message_test.go

internal/adapter/repository/postgres/
└── group_repository.go

internal/adapter/http/
└── group_handler.go

dashboard/src/app/[locale]/groups/
├── page.tsx                    # Group list
└── [groupId]/
    └── page.tsx                # Group details

dashboard/src/components/group/
├── GroupSummary.tsx
├── GroupExpenseList.tsx
├── GroupExpenseForm.tsx
├── GroupMemberList.tsx
├── GroupSettlement.tsx
└── GroupAuditLog.tsx

dashboard/src/infrastructure/repositories/http/
└── groupRepository.ts

dashboard/src/messages/
├── en.json                     # Add group-related translations
├── zh-TW.json
└── ...                         # Other supported locales
```

## Implementation Plan

### Phase 1: Data Layer (Domain + Repository)

**Goal:** Establish data models and storage interfaces

1. Add Domain Models - `internal/domain/group.go`
2. Add Repository Interface - `internal/domain/group_repository.go`
3. Implement Repository - `internal/adapter/repository/postgres/group_repository.go`
4. Database migrations

### Phase 2: Core Business Logic (UseCase)

**Goal:** Implement split calculation and settlement logic

**TDD Approach:** For each feature:
1. Write Gherkin spec (describe behavior)
2. Write failing test (Red)
3. Write minimal implementation to pass (Green)
4. Refactor
5. Run `go test ./internal/usecase/group/... -v` to verify

Features:
1. Group management (create, get, update)
2. Member management (add, remove)
3. Expense management (create, update, delete)
4. Balance calculation (paid/owes/receives per member)
5. Debt calculation (who pays whom how much)
6. Settlement (execute, record to personal ledgers, archive)
7. Audit logging

### Phase 3: API Layer (HTTP Handlers)

**Goal:** Expose REST APIs

1. Add Handler - `internal/adapter/http/group_handler.go`
2. Register routes in `internal/adapter/http/router.go`

### Phase 4: Messenger Integration

**Goal:** Enable chatbot group message support

1. Extend message processing in `internal/usecase/process_message.go`
2. Add `internal/usecase/group/process_message.go` for group logic
3. Update Messenger adapters (LINE, Discord, Telegram) to parse group IDs
4. Handle group commands (enable/disable, continue/new ledger)

### Phase 5: Dashboard Frontend

**Goal:** Build group Dashboard UI

1. Page structure (groups list, group details)
2. Components (summary, expense list, member list, settlement, audit log)
3. API integration
4. i18n translations

### Phase 6: Testing & Verification

**Goal:** Ensure correctness

1. Unit tests - UseCase tests, Repository tests
2. Integration tests - API endpoints, Messenger webhooks
3. E2E tests - Dashboard flows, Chatbot interaction

### Estimated Timeline

| Phase | Item | Estimate |
|-------|------|----------|
| 1 | Data Layer | 2-3 days |
| 2 | Business Logic | 3-4 days |
| 3 | API Layer | 1-2 days |
| 4 | Messenger Integration | 2-3 days |
| 5 | Dashboard Frontend | 4-5 days |
| 6 | Testing & Verification | 2-3 days |
| **Total** | | **14-20 days** |

## Settlement Calculation Example

**Scenario:**
- Ming paid dinner $900
- Hua paid tickets $600
- 3 members split equally

**Per-expense debts:**

Dinner $900 (Ming paid):
- Each owes $300
- Hua owes Ming $300
- Mei owes Ming $300

Tickets $600 (Hua paid):
- Each owes $200
- Ming owes Hua $200
- Mei owes Hua $200

**Settlement (after offsetting):**
- Hua → Ming: $300 - $200 = **$100**
- Mei → Ming: **$300**
- Mei → Hua: **$200**

**Personal ledger records after settlement:**

All three members get:
- Dinner $300
- Tickets $200

This reflects "actual consumption" rather than "debt amounts".
