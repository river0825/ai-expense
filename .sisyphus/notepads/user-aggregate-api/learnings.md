# Learnings

## Frontend Architecture
- The project uses a Clean Architecture approach in the frontend (`domain`, `infrastructure`, `app`).
- Repositories are used to fetch data, often via a `RepositoryFactory`, but can also be instantiated directly (`new HttpAggregateRepository()`).
