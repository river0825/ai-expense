# Change: Redesign Admin Analytics Panel for Revenue and Retention Decisions

## Why
Current metrics endpoints and dashboard coverage focus on generic operational stats and simple API-key protection. The business now needs a decision-grade admin panel centered on revenue and retention, with explicit metric semantics, drill-downs, and independent admin authentication.

## What Changes
- Add a new admin analytics panel capability focused on daily operations and weekly review decisions.
- Extend metrics capability from generic dashboard metrics to revenue/retention KPI contracts (MRR, NRR, GRR, churn, expansion/contraction context).
- Add independent admin authentication capability for admin analytics routes and UI access.
- Add mandatory contract-level and edge-case scenarios (refunds, chargebacks, plan changes, late events, empty cohorts, invalid filters).
- Implement admin frontend only in `frontend/admin` (no feature implementation in `frontend/dashboard`).
- Define rollout guardrails and non-goals to avoid full BI scope creep.

## Impact
- Affected specs:
  - `dashboard-metrics` (MODIFIED)
  - `admin-dashboard` (ADDED)
  - `admin-auth` (ADDED)
- Affected code (planned):
  - Backend metrics/auth routing and use cases under `internal/adapter/http`, `internal/usecase`, `internal/domain`, `internal/adapter/repository/postgresql`
  - Frontend admin experience under `frontend/admin`
  - CI workflow under `.github/workflows/ci.yml`
