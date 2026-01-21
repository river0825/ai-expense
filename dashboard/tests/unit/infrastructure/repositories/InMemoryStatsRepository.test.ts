import { describe, it, expect, beforeEach } from 'vitest';
import { InMemoryStatsRepository } from '@/infrastructure/repositories/in-memory/InMemoryStatsRepository';

describe('InMemoryStatsRepository', () => {
  let repository: InMemoryStatsRepository;

  beforeEach(() => {
    repository = new InMemoryStatsRepository();
  });

  it('should return dashboard stats', async () => {
    const stats = await repository.getDashboardStats();
    
    expect(stats.totalBalance).toBeDefined();
    expect(stats.totalIncome).toBeDefined();
    expect(stats.totalExpenses).toBeDefined();
    
    expect(stats.totalBalance.value).toContain('$');
  });
});
