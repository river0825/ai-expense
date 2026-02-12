## ADDED Requirements

### Requirement: Decision-Focused Admin Overview
The system SHALL provide an admin dashboard overview optimized for daily operations and weekly review decisions.

#### Scenario: Daily operations overview
- **WHEN** admin opens the dashboard default view
- **THEN** system displays top-level revenue/retention KPIs with period-over-period deltas
- **AND** highlights at-risk movement requiring same-day action

#### Scenario: Weekly review overview
- **WHEN** admin switches to weekly review context
- **THEN** system shows trend summaries and cohort/driver views suitable for weekly decision cadence

### Requirement: Admin Frontend Location
The system SHALL implement the admin panel application under `frontend/admin`.

#### Scenario: Admin panel code location
- **WHEN** admin panel features are implemented
- **THEN** routes, components, and tests for the new admin panel are created under `frontend/admin`
- **AND** `frontend/dashboard` is not used for new admin panel feature implementation

### Requirement: Hierarchical Metric Information Architecture
The system SHALL organize admin analytics in a three-level hierarchy to connect outcomes to actions.

#### Scenario: L1 strategic KPIs
- **WHEN** dashboard loads
- **THEN** it displays strategic outcome metrics (for example: MRR, NRR, GRR, churn)

#### Scenario: L2 driver metrics
- **WHEN** admin drills into an L1 KPI
- **THEN** system displays driver metrics and trend decomposition explaining movement

#### Scenario: L3 actionable queues
- **WHEN** admin drills further from L2
- **THEN** system presents actionable entities (for example: at-risk accounts list) with relevant context

### Requirement: Filter and Drill-Down Interactions
The system SHALL provide filter, comparison, and drill-down interactions for decision workflows.

#### Scenario: Period and comparison controls
- **WHEN** admin changes period and comparison settings
- **THEN** KPI cards and charts refresh using the selected context

#### Scenario: Drill-down to concrete records
- **WHEN** admin clicks chart segment or cohort cell
- **THEN** system opens filtered detail view tied to that segment

#### Scenario: Invalid filter selection
- **WHEN** admin selects an unsupported filter combination
- **THEN** system shows a clear validation message and prevents inconsistent query execution

### Requirement: Robust UI States
The system SHALL render clear loading, error, empty, and responsive states for admin analytics.

#### Scenario: Loading state
- **WHEN** analytics data is fetching
- **THEN** dashboard renders deterministic loading placeholders

#### Scenario: Error state
- **WHEN** analytics API request fails
- **THEN** dashboard shows actionable error messaging and retry affordance

#### Scenario: Empty state
- **WHEN** selected range has no data
- **THEN** dashboard renders explicit empty state with non-destructive guidance

#### Scenario: Responsive behavior
- **WHEN** dashboard is viewed on mobile and desktop breakpoints
- **THEN** navigation and metric modules remain usable and legible
