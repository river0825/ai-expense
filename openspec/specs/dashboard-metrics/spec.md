# dashboard-metrics Specification

## Purpose
TBD - created by archiving change add-conversational-expense-tracker. Update Purpose after archive.
## Requirements
### Requirement: Track Daily Active Users (DAU)
The system SHALL track and expose the number of unique active users per day for business metrics.

#### Scenario: DAU endpoint returns daily count
- **WHEN** admin requests `GET /api/metrics/dau`
- **THEN** system returns JSON with daily active users for last 30 days
```json
{
  "data": [
    {"date": "2024-01-16", "active_users": 42},
    {"date": "2024-01-15", "active_users": 38},
    ...
  ],
  "total_30day_active": 150
}
```

#### Scenario: User counted once per day regardless of message count
- **WHEN** user sends 10 messages on 2024-01-16
- **THEN** user is counted as 1 in DAU for that day

#### Scenario: DAU filter by date range
- **WHEN** admin requests `GET /api/metrics/dau?from=2024-01-01&to=2024-01-31`
- **THEN** system returns DAU for specified date range

### Requirement: Track Total Expenses Metrics
The system SHALL aggregate expense data by time period for trend analysis.

#### Scenario: Daily expense total
- **WHEN** admin requests `GET /api/metrics/expenses-summary`
- **THEN** system returns daily totals (sum of all expenses per day)
```json
{
  "data": [
    {"date": "2024-01-16", "total_expense": 1250.50, "count": 15},
    {"date": "2024-01-15", "total_expense": 980.00, "count": 12}
  ],
  "period_average_daily": 1115.25
}
```

#### Scenario: Period aggregation (weekly, monthly)
- **WHEN** admin requests `GET /api/metrics/expenses-summary?period=weekly`
- **THEN** system returns weekly aggregates (sum per week)
- **WHEN** admin requests with `period=monthly`
- **THEN** system returns monthly aggregates

#### Scenario: User count with expenses
- **WHEN** admin requests expense summary
- **THEN** includes count of unique users who had expenses that day

### Requirement: Track Category Trends
The system SHALL provide breakdown of expenses by category for understanding user behavior.

#### Scenario: Category distribution
- **WHEN** admin requests `GET /api/metrics/category-trends`
- **THEN** system returns expense totals grouped by category
```json
{
  "data": [
    {"category": "Food", "total": 5000, "percent": 35, "count": 120},
    {"category": "Transport", "total": 3000, "percent": 21, "count": 60},
    {"category": "Shopping", "total": 4000, "percent": 28, "count": 45},
    {"category": "Entertainment", "total": 2000, "percent": 14, "count": 30},
    {"category": "Other", "total": 300, "percent": 2, "count": 8}
  ],
  "total_expenses": 14300
}
```

#### Scenario: Category trend over time
- **WHEN** admin requests `GET /api/metrics/category-trends?from=2024-01-01&to=2024-01-31`
- **THEN** system returns category breakdown for that period

#### Scenario: Top category by volume and amount
- **WHEN** admin requests category trends
- **THEN** includes which category has highest total amount and highest transaction count

### Requirement: Track User Growth Metrics
The system SHALL provide growth metrics for understanding user acquisition and engagement.

#### Scenario: New users per day
- **WHEN** admin requests `GET /api/metrics/growth`
- **THEN** system returns new user count per day
```json
{
  "new_users_today": 5,
  "new_users_this_week": 32,
  "new_users_this_month": 120,
  "total_users": 450,
  "user_growth_percent_30day": 36.4,
  "data": [
    {"date": "2024-01-16", "new_users": 5},
    {"date": "2024-01-15", "new_users": 4}
  ]
}
```

#### Scenario: Active user retention
- **WHEN** admin requests growth metrics
- **THEN** includes repeat user ratio (users who have used bot multiple days)

#### Scenario: Average expense per user
- **WHEN** admin requests growth metrics
- **THEN** includes avg_expense_per_user (total expenses / unique users)

### Requirement: Metrics Require Authentication
The system SHALL protect metrics endpoints to prevent unauthorized access to business data.

#### Scenario: Metrics endpoint authentication
- **WHEN** client requests metrics without authentication
- **THEN** system returns 401 Unauthorized

#### Scenario: Authenticated metrics access
- **WHEN** authorized admin includes valid API key/token
- **THEN** system returns metrics data

### Requirement: Real-time Metrics Updates
The system SHALL update metrics as new data comes in, with reasonable caching to avoid excessive computation.

#### Scenario: Metrics refresh on expense creation
- **WHEN** user creates expense
- **THEN** DAU, expense summary, category trends are updated
- **AND** updates are reflected in metrics endpoints within 5 seconds

#### Scenario: Caching for performance
- **WHEN** multiple requests for same metrics within short time
- **THEN** system uses cached data to avoid recalculation
- **AND** cache is invalidated when new expense/user created

