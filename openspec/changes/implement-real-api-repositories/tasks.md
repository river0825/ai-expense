# Implement Real API Repositories Tasks

- [ ] Setup Next.js API Routes
    - [ ] Create `src/app/api/transactions/route.ts` (GET).
    - [ ] Create `src/app/api/stats/route.ts` (GET).
- [ ] Implement API Repositories
    - [ ] Create `src/infrastructure/repositories/api/ApiTransactionRepository.ts`.
    - [ ] Create `src/infrastructure/repositories/api/ApiStatsRepository.ts`.
- [ ] Update Factory
    - [ ] Update `src/infrastructure/RepositoryFactory.ts` to return API repositories.
- [ ] Verify
    - [ ] Verify Dashboard loads data from API (check Network tab).
    - [ ] Verify error handling (e.g. 500 status).
