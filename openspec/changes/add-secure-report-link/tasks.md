# Tasks: Add Secure Report Link

## Phase 1: Backend Implementation
- [ ] 1.1 Implement JWT generation utility (or extend existing auth).
- [ ] 1.2 Create `GenerateReportLinkUseCase`.
- [ ] 1.3 Update `ProcessMessageUseCase` to handle "view report" intent.
- [ ] 1.4 Create `GetReportSummaryUseCase` for the report data.
- [ ] 1.5 Create API endpoint `GET /api/reports/summary` with token auth.

## Phase 2: Frontend Implementation
- [ ] 2.1 Create `app/report/page.tsx`.
- [ ] 2.2 Implement API client for fetching report data.
- [ ] 2.3 Build UI components (Summary Card, Category Chart, Expense List).
- [ ] 2.4 Handle loading and error states (Expired Token).

## Phase 3: Verification
- [ ] 3.1 Test Intent Detection (Unit Test).
- [ ] 3.2 Test Token Generation & Validation (Integration Test).
- [ ] 3.3 Verify full flow manually.
