import { TransactionRepository } from '@/domain/repositories/TransactionRepository';
import { StatsRepository } from '@/domain/repositories/StatsRepository';
import { ReportRepository } from '@/domain/repositories/ReportRepository';
import { ExpenseRepository } from '@/domain/repositories/ExpenseRepository';
import { InMemoryTransactionRepository } from './repositories/in-memory/InMemoryTransactionRepository';
import { InMemoryStatsRepository } from './repositories/in-memory/InMemoryStatsRepository';
import { HttpReportRepository } from './repositories/http/HttpReportRepository';
import { HttpExpenseRepository } from './repositories/http/HttpExpenseRepository';
import { UserRepository } from '@/domain/repositories/UserRepository';
import { CurrencyRepository } from '@/domain/repositories/CurrencyRepository';
import { HttpUserRepository } from './repositories/http/HttpUserRepository';
import { HttpCurrencyRepository } from './repositories/http/HttpCurrencyRepository';
import { CategoryRepository } from '@/domain/repositories/CategoryRepository';
import { HttpCategoryRepository } from './repositories/http/HttpCategoryRepository';

class RepositoryFactory {
  private static transactionRepository: TransactionRepository;
  private static statsRepository: StatsRepository;
  private static reportRepository: ReportRepository;
  private static expenseRepository: ExpenseRepository;
  private static userRepository: UserRepository;
  private static currencyRepository: CurrencyRepository;
  private static categoryRepository: CategoryRepository;

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

  static getUserRepository(): UserRepository {
    if (!this.userRepository) {
      this.userRepository = new HttpUserRepository();
    }
    return this.userRepository;
  }

  static getCurrencyRepository(): CurrencyRepository {
    if (!this.currencyRepository) {
      this.currencyRepository = new HttpCurrencyRepository();
    }
    return this.currencyRepository;
  }

  static getCategoryRepository(): CategoryRepository {
    if (!this.categoryRepository) {
      this.categoryRepository = new HttpCategoryRepository();
    }
    return this.categoryRepository;
  }
}

export default RepositoryFactory;
