## ADDED Requirements

### Requirement: Interactive Dashboard
The system SHALL provide an interactive dashboard for users to view their expense reports.

#### Scenario: Dashboard Initialization
Given a user lands on the `/reports` page
When the page loads
Then it checks for a valid authentication token (Cookie or URL param)
And if valid, it automatically fetches report data for the default range (e.g., Current Month)
And it displays the "Total Expense" card
And it displays the "Expenses by Category" chart
And it displays the "Recent Expenses" list

### Requirement: Data Filtering
The dashboard SHALL allow users to filter expense data by date range.

#### Scenario: Date Range Selection
Given the user is viewing the report dashboard
When the user changes the date range using the picker
Then the dashboard immediately fetches new data for the selected range
And the Chart, Stats, and List update to reflect the new data

### Requirement: Responsive Design
The dashboard SHALL be responsive and usable on mobile devices.

#### Scenario: Mobile Responsiveness
Given a user accesses the dashboard on a mobile device
When the page renders
Then the layout adapts to a single column view
And the charts remain legible
And the expense list is scrollable
