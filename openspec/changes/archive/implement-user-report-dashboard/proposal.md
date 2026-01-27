# Change Proposal: Implement User Report Dashboard

## Problem
The current report link provides a static snapshot or basic view. Users need an interactive dashboard to explore their expenses, filter by date ranges (monthly, weekly), and view specific metrics without logging into the main admin dashboard.

## Proposed Solution
Create a new, dedicated user-facing dashboard accessible via a secure, temporary short link. This dashboard will be read-only and scoped to the user's data.

## Key Features
1.  **Secure Access**:
    *   Short links (valid for 5 minutes) redirect to the dashboard.
    *   JWT authentication (set in cookie upon redirect).
2.  **Interactive UI**:
    *   Date range picker (Monthly, Weekly presets).
    *   Key Metrics: Total Expense (no Income/Balance yet).
    *   Visualizations: Expense Chart grouped by category.
    *   Data: Detailed expense list.
3.  **API Support**:
    *   Endpoints to fetch report data based on date range and user token.

## Capabilities
-   **User Dashboard**: The frontend interface for viewing reports.
-   **Reporting**: Backend logic for aggregating expense data dynamically.
-   **Short Link**: Mechanism for generating and validating temporary access links.

## Success Criteria
-   User can request a report via chat.
-   User receives a short link (e.g., `/r/xyz123`).
-   Link redirects to dashboard and sets auth cookie.
-   Dashboard loads and displays correct expense data for the default range (last 30 days).
-   User can change date range and see updated data.
