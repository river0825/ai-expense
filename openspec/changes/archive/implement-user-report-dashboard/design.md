# Design: User Report Dashboard

## Architecture

### 1. Access Flow
1.  **User Request**: User types "report" in chat.
2.  **Link Generation**:
    *   Backend generates a JWT `report_token` (valid for 15 mins).
    *   Backend generates a random Short ID (e.g., `abc123`).
    *   Backend stores `ShortID -> {Token, Expiry}` in `short_links` table (Expiry: 5 mins).
    *   Bot sends `https://api.aiexpense.com/r/abc123` to user.
3.  **Redirection**:
    *   User clicks link.
    *   `GET /r/{id}` endpoint validates ID.
    *   Server responds with `302 Redirect` to `https://dashboard.aiexpense.com/reports`.
    *   **Crucial**: The redirect response includes `Set-Cookie: report_token=...; HttpOnly; Path=/`.
    *   Alternatively (if cross-domain issues arise): Redirect to `https://dashboard.aiexpense.com/auth?token=...` which sets the cookie client-side or server-side. *Decision: Use query param for simplicity in this iteration, as Dashboard and API might be on different domains/ports in dev.* -> *Correction per user request: "redirect it and set JWT cookie". We will attempt HTTP redirect with Set-Cookie if on same domain, or redirect with token in query param and let frontend set cookie.*

### 2. Frontend (User Dashboard)
*   **Route**: `/reports`
*   **Auth**: Middleware or Page logic checks for `report_token` cookie.
*   **Components**:
    *   `DateRangePicker`: Allows custom range or presets (This Month, Last Month, This Week).
    *   `StatsCard`: Shows Total Expense.
    *   `CategoryChart`: Pie/Donut chart of expenses by category.
    *   `ExpenseList`: Table/List of individual transactions.
*   **Data Fetching**:
    *   Calls `GET /api/reports/summary?start_date=...&end_date=...`
    *   Authorization header: `Bearer <token>` (from cookie).

### 3. Backend (API)
*   **Endpoint**: `GET /api/reports/summary`
*   **Params**: `start_date`, `end_date`.
*   **Auth Middleware**: Validates `report_token` JWT. Extracts `user_id`.
*   **Response**:
    ```json
    {
      "total_expense": 1250.50,
      "categories": [
        {"name": "Food", "amount": 450.00, "color": "#..."},
        ...
      ],
      "expenses": [
        {"id": "...", "date": "2023-10-27", "description": "Lunch", "amount": 15.00, "category": "Food"},
        ...
      ]
    }
    ```

## Database Changes
*   **Table**: `short_links`
    *   `id` (PK, string, 6-8 chars)
    *   `target_token` (text, JWT)
    *   `expires_at` (timestamp)
    *   `created_at` (timestamp)

## Security Considerations
*   Short links expire quickly (5 mins) to prevent replay/leakage.
*   JWT expires (15 mins) to limit session duration.
*   Dashboard is read-only.
*   Data access is strictly scoped to the `user_id` in the JWT.
