## Context
The current product has a metrics capability and dashboard UI, but it is oriented toward generic operational metrics and API-key-based endpoint access. The requested redesign requires a dedicated admin analytics experience optimized for business decisions (Revenue and Retention), with independent admin authentication and explicit metric semantics.

## Goals / Non-Goals
- Goals:
  - Define a decision-grade admin analytics capability for daily operations and weekly review.
  - Standardize revenue/retention metric contracts and edge-case semantics.
  - Require independent admin authentication and explicit authorization boundaries.
  - Ensure testable API/UI behavior and CI enforcement.
- Non-Goals:
  - Full BI query builder, scheduled reporting pipeline, or forecasting engine.
  - Broad enterprise IAM rollout beyond admin analytics needs for this change.

## Decisions
- Decision: Introduce a new admin analytics capability while modifying existing `dashboard-metrics` requirements.
  - Why: Existing metrics capability already owns API metric semantics; admin experience requires additional behavior and constraints.
- Decision: Define new `admin-auth` capability for independent authentication requirements.
  - Why: Existing metric access via static `X-API-Key` is insufficient for long-term admin session and authorization requirements.
- Decision: Specify formulas and edge-case handling at spec level (MRR/NRR/GRR/churn).
  - Why: Prevents implementation ambiguity and test drift.

## Risks / Trade-offs
- Risk: Ambiguous metric semantics can lead to conflicting backend/frontend assumptions.
  - Mitigation: Contract-level requirement language with formula and scenario coverage.
- Risk: Scope creep into broad analytics platform work.
  - Mitigation: Explicit non-goals and bounded requirement set.
- Risk: Security regressions from mixed auth approaches.
  - Mitigation: Dedicated `admin-auth` capability and explicit negative scenarios.

## Migration Plan
1. Add OpenSpec deltas for `dashboard-metrics`, `admin-dashboard`, and `admin-auth`.
2. Validate with strict mode.
3. Implement backend contracts/auth and frontend admin panel in staged tasks.
4. Archive change after deployment with spec updates.

## Open Questions
- Whether admin auth must support MFA in V1 or only independent login/session boundary.
- Whether finance-grade revenue source will be internal transaction ledger only or external billing integration.
