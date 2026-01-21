# Implement Real API Repositories

## Summary
Implement `ApiTransactionRepository` and `ApiStatsRepository` to fetch data from real API endpoints, replacing the in-memory implementations. This also involves setting up Next.js API routes to serve as the backend.

## Context
Currently, the application uses `InMemoryRepository` with hardcoded data. To move towards a production-ready application, we need to fetch data from an API.

## Goals
1. Create Next.js API Routes (`/api/transactions`, `/api/stats`) to serve data.
2. Implement `ApiTransactionRepository` and `ApiStatsRepository` using `axios`.
3. Update `RepositoryFactory` to provide the API repositories.
4. Ensure the Dashboard works with the requested "real" data.

## Non-Goals
- Connecting to an external database (Postgres/MySQL) - we will still serve mock data *from* the API routes for now, as the focus is on the *request layer*.

## Plan
1. **API Routes**: Create `src/app/api/transactions/route.ts` and `src/app/api/stats/route.ts`.
2. **Repositories**: Implement `ApiTransactionRepository` and `ApiStatsRepository` in `src/infrastructure/repositories/api/`.
3. **Integration**: Switch `RepositoryFactory` to use the new API repositories.
4. **Verification**: Verify data loading and error handling.
