import { TransactionRepository } from '@/domain/repositories/TransactionRepository';
import { StatsRepository } from '@/domain/repositories/StatsRepository';
import { InMemoryTransactionRepository } from './repositories/in-memory/InMemoryTransactionRepository';
import { InMemoryStatsRepository } from './repositories/in-memory/InMemoryStatsRepository';

class RepositoryFactory {
  private static transactionRepository: TransactionRepository;
  private static statsRepository: StatsRepository;

  static getTransactionRepository(): TransactionRepository {
    if (!this.transactionRepository) {
      this.transactionRepository = new InMemoryTransactionRepository();
    }
    return this.transactionRepository;
  }

  static getStatsRepository(): StatsRepository {
    if (!this.statsRepository) {
      this.statsRepository = new InMemoryStatsRepository();
    }
    return this.statsRepository;
  }
}

export default RepositoryFactory;
