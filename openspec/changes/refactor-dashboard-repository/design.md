# Repository Pattern Design

## Architecture

We will implement the **Repository Pattern** to decouple the domain layer (UI) from the data layer.

### Interfaces
We will define strict TypeScript interfaces in `src/domain/repositories/`:

```typescript
export interface Transaction {
  id: string;
  name: string;
  category: string;
  date: string;
  amount: number;
  currency: string;
  status: 'Completed' | 'Pending';
}

export interface TransactionRepository {
  getRecentTransactions(limit: number): Promise<Transaction[]>;
  getAllTransactions(): Promise<Transaction[]>;
}

export interface Stats {
  totalBalance: number;
  totalIncome: number;
  totalExpenses: number;
  // ... other stats
}

export interface StatsRepository {
  getDashboardStats(): Promise<Stats>;
}
```

### Implementations

We will create a `src/infrastructure/repositories/` directory for implementations:

1.  **InMemory**: `src/infrastructure/repositories/in-memory/`
    *   Uses static arrays to simulate a database.
    *   Simulates network latency with `setTimeout` if needed.

### Dependency Injection

For Next.js App Router:
*   **Server Components**: We can instantiate the repository directly in server components or use a service locator pattern if we need to switch implementations dynamically based on environment variables.
*   **Client Components**: Data should be passed from Server Components as props, or fetched via Route Handlers which use the Repository.

Given the current setup (Client Component `page.tsx` using `useTranslations`), we might need to:
1.  Keep `page.tsx` as a Client Component but fetch data in a parent Server Component (`layout.tsx` or a new wrapper).
2.  Or, for simplicity in this refactor, just instantiate the repository in `page.tsx` (or a hook) until we move to full Server Actions/Components data fetching. 

*Decision*: We will expose a `RepositoryFactory` that returns the configured implementation (In-Memory for now).

## Directory Structure Changes

```
src/
  domain/
    models/
    repositories/
  infrastructure/
    repositories/
      in-memory/
  app/
    ...
```
