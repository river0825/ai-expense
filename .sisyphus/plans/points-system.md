# Points System & Group Chat MVP

## TL;DR

> **Quick Summary**: Implement a prepaid points system (no subscription) with collaborative group features to drive viral growth. New users get 500 points (sunk cost strategy); advanced group features consume more points.
>
> **Deliverables**:
> - New Domain Models: `Wallet`, `PointLog`, `Group`, `GroupMember`
> - Group Logic: Create/Join/Bind External Chat (Line/Telegram)
> - Points Logic: Deduct points (1/expense, 50/trip), Mock Top-up, Invite Rewards
> - Enforcement: Read-only mode when points <= 0
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Database Migration → Wallet/Group Core → Chat Integration

---

## Context

### Original Request
User wants a profitable model based on points (prepaid) rather than subscription.
Strategy:
1.  **Sunk Cost**: High initial grant (500 pts) to build habit.
2.  **Viral Loop**: Invite rewards via group chats.
3.  **Monetization**: Advanced group features (splitting/multi-currency) cost more.
4.  **No Lock-out**: When points=0, read-only mode (data visible but frozen).

### Technical Decisions
- **Group Binding**: External Chat ID (Line/Telegram) mapped to internal `group_id`.
- **Deduction Rule**: Group expenses deduct from the **Group Owner's** wallet.
- **Payment**: Mock implementation (admin endpoint / UI button) for MVP.
- **Architecture**: New `WalletService` and `GroupService` layers.

### Metis Review & Guardrails
- **Guardrail**: Prevent negative balance (Atomic check-and-deduct).
- **Guardrail**: Handle "Owner out of points" scenario (Reject group expense with explicit error).
- **Scope Limit**: No real payment gateway integration (Mock only).
- **Scope Limit**: No complex "Group Pool" logic (Owner pays all for MVP).

---

## Work Objectives

### Core Objective
Enable multi-user group expense tracking powered by a prepaid points economy.

### Concrete Deliverables
- [ ] Database migrations for `wallets`, `point_logs`, `groups`, `group_members`
- [ ] Backend API for Wallet (Balance, History, Mock Top-up)
- [ ] Backend API for Groups (Create, Join, Invite Link)
- [ ] Middleware/Service logic to intercept Expense Creation for point deduction
- [ ] Logic to auto-create Wallet + 500 pts for new users

### Definition of Done
- [ ] New user signup = Balance 500
- [ ] Create expense (Personal) = Balance -1
- [ ] Create expense (Group) = Owner Balance -1
- [ ] Balance 0 = "Create" returns 402 Payment Required
- [ ] Invite flow = Referrer +30 pts, Referee +30 pts

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES (Go testing)
- **Automated tests**: YES (TDD)
- **Framework**: `go test` standard library

### Agent-Executed QA Scenarios (MANDATORY)

#### Scenario 1: New User Sunk Cost Flow
**Tool**: Bash (curl)
1.  **Signup**: Create new user via API
2.  **Assert**: Wallet created with 500 points
3.  **Create Expense**: Create 1 personal expense
4.  **Assert**: Balance becomes 499
5.  **Log Check**: Verify `point_logs` entry exists (reason: "expense_creation")

#### Scenario 2: Group Expense & Deduction
**Tool**: Bash (curl)
1.  **Setup**: User A (Owner, 500 pts) creates Group X
2.  **Join**: User B (0 pts) joins Group X
3.  **Action**: User B creates expense in Group X
4.  **Assert**: Expense created successfully
5.  **Assert**: User A balance = 499 (Deducted from Owner)
6.  **Assert**: User B balance = 0 (Unchanged)

#### Scenario 3: Bankrupt Enforcement (Read-Only Mode)
**Tool**: Bash (curl)
1.  **Setup**: User C has 0 points
2.  **Action**: User C attempts to create expense
3.  **Assert**: HTTP 402 Payment Required
4.  **Action**: User C attempts to GET expenses
5.  **Assert**: HTTP 200 OK (Read-only works)

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Core Data):
├── Task 1: DB Migrations & Domain Models
└── Task 2: WalletService (Logic only)

Wave 2 (Feature Implementation):
├── Task 3: Point Deduction Interceptor (Expense UseCase)
├── Task 4: Group Management API
└── Task 5: Mock Payment & Initial Grant Logic

Wave 3 (Integration & Viral):
├── Task 6: External Chat Binding Logic
└── Task 7: Invite System & Rewards
```

---

## TODOs

- [ ] 1. **Database Schema & Domain Models**
    **What to do**:
    - Create SQL migration for:
        - `wallets` (user_id, balance, updated_at)
        - `point_logs` (id, wallet_id, amount, reason, reference_id, created_at)
        - `groups` (id, owner_id, name, external_id, created_at)
        - `group_members` (group_id, user_id, joined_at)
    - Update `Expense` model to add `GroupID` (nullable)
    - Define Structs in `internal/domain/models.go`
    
    **Parallelization**:
    - **Can Run In Parallel**: YES
    - **Wave**: 1
    
    **References**:
    - `internal/domain/models.go` (Existing models)
    - `migrations/` (Migration pattern)

    **Acceptance Criteria**:
    - [ ] Migration runs successfully (`make migrate-up`)
    - [ ] Tables created with correct foreign keys
    - [ ] `Expense` struct has `GroupID` field

- [ ] 2. **Wallet Service (Core Logic)**
    **What to do**:
    - Implement `WalletRepository` (GetBalance, Add, Deduct)
    - Implement `WalletService`:
        - `Deduct(userID, amount, reason)` -> returns error if insufficient
        - `Add(userID, amount, reason)`
    - Ensure atomic updates (SELECT FOR UPDATE or UPDATE ... WHERE balance >= amount)
    
    **Parallelization**:
    - **Can Run In Parallel**: YES
    - **Wave**: 1
    
    **References**:
    - `internal/domain/repositories.go` (Repo interfaces)
    
    **Acceptance Criteria**:
    - [ ] Unit tests for Atomic Deduction (concurrent requests shouldn't go negative)
    - [ ] Unit tests for Insufficient Funds error

- [ ] 3. **Expense Deduction Interceptor**
    **What to do**:
    - Modify `CreateExpenseUseCase`:
        - Before saving, call `WalletService.Deduct(1)`
        - If GroupID present, look up GroupOwner, call `WalletService.Deduct(1)` from Owner
        - If error (402), abort creation
    - Add "Read-Only" check (if balance <= 0, allow List but block Create/Update)
    
    **Parallelization**:
    - **Can Run In Parallel**: NO (Depends on 1 & 2)
    - **Wave**: 2
    
    **References**:
    - `internal/usecase/create_expense.go`
    
    **Acceptance Criteria**:
    - [ ] QA Scenario 1 (Personal deduction)
    - [ ] QA Scenario 2 (Group Owner deduction)
    - [ ] QA Scenario 3 (Bankrupt block)

- [ ] 4. **Group Management API**
    **What to do**:
    - Implement `GroupUseCase`:
        - `CreateGroup(ownerID, name)`
        - `JoinGroup(userID, groupID)`
        - `GetUserGroups(userID)`
    - Add API Handlers for Groups
    
    **Parallelization**:
    - **Can Run In Parallel**: YES
    - **Wave**: 2
    
    **References**:
    - `internal/adapter/http/handler.go`
    
    **Acceptance Criteria**:
    - [ ] API `POST /groups` creates group
    - [ ] API `POST /groups/:id/join` adds member

- [ ] 5. **New User Grant & Mock Payment**
    **What to do**:
    - Modify User Registration (or first login) to call `WalletService.Add(500, "welcome_bonus")`
    - Create `POST /api/wallet/mock-topup`:
        - Admin/Debug endpoint to add points
        - Body: `{ amount: 100 }`
    
    **Parallelization**:
    - **Can Run In Parallel**: YES
    - **Wave**: 2
    
    **Acceptance Criteria**:
    - [ ] New user automatically has 500 points
    - [ ] Mock endpoint increases balance

- [ ] 6. **External Chat Binding**
    **What to do**:
    - Modify `ParseConversationUseCase`:
        - Extract `SourceID` (Line Group ID) from metadata
        - Check if `groups` table has `external_id == SourceID`
        - If yes, assign `GroupID` to the expense automatically
    - If no internal group exists for this external ID:
        - (Optional for MVP) Auto-create group? Or prompt user to link?
        - **Decision**: For MVP, if SourceID exists but no link, treat as Personal.
    
    **Parallelization**:
    - **Can Run In Parallel**: YES
    - **Wave**: 3
    
    **References**:
    - `internal/usecase/parse_conversation.go`
    
    **Acceptance Criteria**:
    - [ ] Expense from bound Line Group automatically gets internal `GroupID`
    - [ ] Deduction logic correctly charges the Group Owner

- [ ] 7. **Viral Invite System**
    **What to do**:
    - Create `InviteCode` table or logic (hash of userID)
    - API `POST /invite/claim`:
        - Input: `code`
        - Action: Add 30 pts to Referrer, 30 pts to User
        - Limit: User can only claim once
    
    **Parallelization**:
    - **Can Run In Parallel**: YES
    - **Wave**: 3
    
    **Acceptance Criteria**:
    - [ ] Claiming code updates both balances
    - [ ] Duplicate claim blocked
