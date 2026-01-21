import { describe, it, expect, beforeEach } from 'vitest';
import { InMemoryTransactionRepository } from '@/infrastructure/repositories/in-memory/InMemoryTransactionRepository';
import { Transaction } from '@/domain/models/Transaction';

describe('InMemoryTransactionRepository', () => {
  let repository: InMemoryTransactionRepository;

  beforeEach(() => {
    repository = new InMemoryTransactionRepository();
  });

  it('should return recent transactions (limit)', async () => {
    const transactions = await repository.getRecentTransactions(3);
    expect(transactions.length).toBe(3);
    expect(transactions[0].id).toBe(1); // Assuming ID 1 is the most recent in mock data
  });

  it('should return all transactions', async () => {
    const transactions = await repository.getAllTransactions();
    expect(transactions.length).toBeGreaterThan(0);
  });
});
