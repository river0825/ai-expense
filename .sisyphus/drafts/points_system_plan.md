# Draft: Points System & Group Features

## Requirements (confirmed)
- **Business Model**: Prepaid Points System (No Subscription).
  - 1 Point = 1 Record (Basic Expense).
  - Higher cost for advanced features (Group Splitting, Multi-currency).
  - **Strategy**: "Sunk Cost" / High Initial Grant.
    - New users get **500 Points** (approx 2-3 months usage).
    - When points = 0, user can VIEW but NOT CREATE (Read-only mode).
- **Viral Mechanism**: Invite rewards (User gets points for inviting others).
- **Core Features**:
  - Points Wallet (Purchase/Deduct/Log).
  - Group Management (Create/Join/List).
  - Collaborative Expense Tracking (Expenses belong to Group).
  - "Group Chat" context: **Bind to External Messenger (LINE/Telegram)**.
- **Tech Decisions**:
  - **Group Chat**: External Messenger binding (store `ExternalGroupID`).
  - **Payment**: Mock implementation for MVP.
  - **Deduction Rule**: Deduct from **Group Owner** (or Group Pool logic if Owner empty).

## Technical Gap Analysis
- **Missing Models**:
  - `User.PointsBalance` (need migration).
  - `PointTransaction` (audit log: type, amount, related_id).
  - `Group` (id, name, created_by).
  - `GroupMember` (group_id, user_id, role).
  - `Expense.GroupID` (nullable, for group expenses).
- **Missing Logic**:
  - `WalletService`: Handle top-up, deduction, insufficient funds check.
  - `GroupService`: Handle join links, invite codes.
  - `ExpenseService`: Update to check User Balance before creation.

## Open Questions
- **Payment Gateway**: How to implement "buying points"? (Mock first or real Stripe/Line Pay?)
- **Chat Integration**: Is "Group Chat" a new internal chat system, or mapping external messenger (Line/Telegram) Group IDs to our system?
- **Initial Grant**: How many points for new users? (Proposed: 100)

## Scope Boundaries
- **INCLUDE**: 
  - Backend API for Wallet & Groups.
  - Frontend UI for Wallet (Balance, Top-up Mock) & Groups.
  - Logic to deduct points on actions.
- **EXCLUDE**:
  - Real Payment Gateway integration (Mock only for MVP).
  - Real-time internal chat (Assume expenses come from external source or simple input form first).
