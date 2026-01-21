import { StatsRepository } from '@/domain/repositories/StatsRepository';
import { DashboardStats } from '@/domain/models/Stats';

const MOCK_STATS: DashboardStats = {
  totalBalance: {
    value: '$24,562.00',
    change: '+2.9%',
    isPositive: true
  },
  totalIncome: {
    value: '$8,240.50',
    change: '+12.999%',
    isPositive: true
  },
  totalExpenses: {
    value: '$3,820.00',
    change: '-4.1%',
    isPositive: false
  }
};

export class InMemoryStatsRepository implements StatsRepository {
  async getDashboardStats(): Promise<DashboardStats> {
    return MOCK_STATS;
  }
}
