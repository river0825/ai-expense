# Refactor Dashboard Repository Layer

## Summary
Refactor the Dashboard application to introduce a Repository Layer for data access. This separates data retrieval logic from UI components, enabling easier testing and future integration with real APIs.

## Context
Currently, the Dashboard uses hardcoded mock data for Stats and Transactions directly in the React components. As the application grows and integrates with a backend, we need a consistent abstraction for data access.

## Goals
1. Define a clear `Repository` interface for data entities (Transactions, Stats).
2. **Setup testing infrastructure (Vitest)** to enable Test-Driven Development.
3. Implement an `InMemoryRepository` for development and testing.
4. Prepare the architecture for a future `ApiRepository` or `LocalStorageRepository`.

## Non-Goals
- Implementing the actual backend API integration in this change (this will be a follow-up).
- Persisting in-memory changes across reloads (in-memory will be ephemeral).

## Plan
1. **Setup Test Infrastructure**: Install Vitest and configure it for the Next.js project.
2. **Define Interfaces**: Create TypeScript interfaces for `TransactionRepository` and `StatsRepository`.
3. **TDD Implementation**:
    - Write failing tests for `InMemoryTransactionRepository`.
    - Implement the repository to pass tests.
    - Write failing tests for `InMemoryStatsRepository`.
    - Implement the repository to pass tests.
4. **Integration**: Inject the repositories into the Next.js application and update `Dashboard` page to fetch data from them.
