import { TransactionRepository } from '@/domain/repositories/TransactionRepository';
import { StatsRepository } from '@/domain/repositories/StatsRepository';
import { ReportRepository } from '@/domain/repositories/ReportRepository';
import { ExpenseRepository } from '@/domain/repositories/ExpenseRepository';
import { InMemoryTransactionRepository } from './repositories/in-memory/InMemoryTransactionRepository';
import { InMemoryStatsRepository } from './repositories/in-memory/InMemoryStatsRepository';
import { HttpReportRepository } from './repositories/http/HttpReportRepository';
import { HttpExpenseRepository } from './repositories/http/HttpExpenseRepository';

class RepositoryFactory {
  private static transactionRepository: TransactionRepository;
  private static statsRepository: StatsRepository;
  private static reportRepository: ReportRepository;
  private static expenseRepository: ExpenseRepository;

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

  static getReportRepository(): ReportRepository {
    if (!this.reportRepository) {
      this.reportRepository = new HttpReportRepository();
    }
    return this.reportRepository;
  }

  static getExpenseRepository(): ExpenseRepository {
    if (!this.expenseRepository) {
      this.expenseRepository = new HttpExpenseRepository();
    }
    return this.expenseRepository;
  }
}

export default RepositoryFactory;
