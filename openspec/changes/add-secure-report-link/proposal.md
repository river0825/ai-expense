# Change Proposal: Secure Expense Report Link

## Problem
Users currently interact with the bot solely via text. They need a way to visualize their expenses (charts, summaries) without logging into a full dashboard with username/password, as the primary interface is the chat app.

## Proposed Solution
Add a "View Report" intent to the chatbot. When triggered, the bot generates a secure, time-limited link (e.g., `https://dashboard.aiexpense.com/report?token=...`). Clicking this link opens a read-only report page in the Next.js dashboard, authenticated via the token.

## Key Changes
1.  **Backend**:
    -   Update `ProcessMessageUseCase` to detect "view report" intent (or similar keywords).
    -   Create `GenerateReportLinkUseCase` to generate a signed JWT/token.
    -   Create `GetReportDataUseCase` (or reuse existing) exposed via a new API endpoint `/api/reports/view` that accepts the token.
2.  **Frontend**:
    -   Create a new page `app/report/page.tsx` in the dashboard.
    -   Implement token validation and data fetching.
    -   Design a mobile-friendly report UI (Summary, Pie Chart, Recent List) using UI/UX Pro Max principles.

## Security Considerations
-   **Token**: JWT or HMAC-signed token.
-   **Expiration**: Short-lived (e.g., 15-30 minutes).
-   **Scope**: Read-only access to the user's expense report.
-   **Transport**: HTTPS only.

## Impact
-   **Users**: Enhanced experience, visual insights.
-   **Architecture**: New auth mechanism (token-based query param) for specific pages.

## Status
-   [ ] Proposed
-   [ ] Approved
-   [ ] Implemented
