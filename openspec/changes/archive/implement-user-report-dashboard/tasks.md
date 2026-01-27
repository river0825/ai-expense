# Tasks: Implement User Report Dashboard

## Phase 1: Backend - Core Logic & API
- [ ] 1.1 Create `ShortLink` domain model and repository interface.
- [ ] 1.2 Implement `ShortLinkRepository` (Postgres/SQLite).
- [ ] 1.3 Create `GenerateReportLinkUseCase` to generate JWT and Short Link.
- [ ] 1.4 Implement `GET /r/{id}` endpoint to handle redirection and cookie setting.
- [ ] 1.5 Update `GetReportSummary` API to accept `start_date` and `end_date` query parameters.
- [ ] 1.6 Update `GetReportSummary` API to validate JWT from Cookie/Header.

## Phase 2: Frontend - Dashboard UI
- [ ] 2.1 Create `/reports` page layout (separate from Admin layout if needed).
- [ ] 2.2 Implement `DateRangePicker` component.
- [ ] 2.3 Create `ReportStats` component (Total Expense).
- [ ] 2.4 Create `CategoryChart` component using Recharts (or similar).
- [ ] 2.5 Create `ExpenseList` component.
- [ ] 2.6 Implement API integration: Fetch data based on selected date range.
- [ ] 2.7 Implement Token handling: Extract from URL/Cookie and persist.

## Phase 3: Integration & Polish
- [ ] 3.1 Verify "Report" command in Chatbot generates valid short link.
- [ ] 3.2 Verify clicking link redirects and logs user into dashboard.
- [ ] 3.3 Verify date filtering updates charts and lists correctly.
- [ ] 3.4 Style UI to match "UI/UX Pro Max" standards (clean, responsive, dark mode compatible).
