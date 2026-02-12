## MODIFIED Requirements

### Requirement: Track User Growth Metrics
The system SHALL provide growth and retention metrics for understanding user acquisition, engagement, and revenue preservation.

#### Scenario: New users per day
- **WHEN** admin requests `GET /api/metrics/growth`
- **THEN** system returns new user count per day and aggregate rollups for today/week/month

#### Scenario: Active user retention
- **WHEN** admin requests growth metrics
- **THEN** system includes repeat user ratio and clearly labels retention window used for calculation

#### Scenario: Average expense per user
- **WHEN** admin requests growth metrics
- **THEN** system includes `avg_expense_per_user` with explicit denominator definition

#### Scenario: Revenue retention metrics are defined with explicit formulas
- **WHEN** admin requests revenue-retention summary for a period
- **THEN** system computes and returns `nrr` as `(starting_mrr + expansion_mrr - contraction_mrr - churned_mrr) / starting_mrr`
- **AND** system computes and returns `grr` as `(starting_mrr - contraction_mrr - churned_mrr) / starting_mrr`
- **AND** system returns formula inputs in the payload for auditability

### Requirement: Metrics Require Authentication
The system SHALL protect metrics endpoints to prevent unauthorized access to business data using admin-authenticated access.

#### Scenario: Metrics endpoint authentication
- **WHEN** client requests protected analytics endpoint without valid admin authentication
- **THEN** system returns `401 Unauthorized`

#### Scenario: Authenticated metrics access
- **WHEN** authorized admin uses valid auth credentials/session
- **THEN** system returns metrics data

#### Scenario: Forbidden access for insufficient authorization
- **WHEN** authenticated principal lacks admin analytics permission
- **THEN** system returns `403 Forbidden`

## ADDED Requirements

### Requirement: Revenue and Retention Metric Semantics
The system SHALL define and expose revenue and retention metrics with explicit business semantics to support decision-making.

#### Scenario: MRR definition excludes non-recurring values
- **WHEN** system computes MRR for a period
- **THEN** only recurring subscription revenue is included
- **AND** one-time charges, taxes, and non-recurring adjustments are excluded unless explicitly configured

#### Scenario: Churn components are represented explicitly
- **WHEN** system returns revenue-retention metrics
- **THEN** payload includes `churned_mrr`, `contraction_mrr`, and `expansion_mrr` as separate fields

#### Scenario: Edge events are handled deterministically
- **WHEN** refunds, chargebacks, or late-arriving events affect the selected period
- **THEN** system applies documented adjustment rules and includes source-period attribution in response metadata
