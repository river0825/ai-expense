import { TransactionRepository } from '@/domain/repositories/TransactionRepository';
import { Transaction } from '@/domain/models/Transaction';

const MOCK_TRANSACTIONS: Transaction[] = [
  { id: 1, name: 'AWS Infrastructure', category: 'Services', date: 'Oct 24, 2024', amount: '-$240.00', status: 'Completed', account: 'Business Card' },
  { id: 2, name: 'Stripe Payment', category: 'Income', date: 'Oct 23, 2024', amount: '+$1,250.00', status: 'Completed', account: 'Business Checking' },
  { id: 3, name: 'Slack Subscription', category: 'Software', date: 'Oct 22, 2024', amount: '-$12.00', status: 'Pending', account: 'Business Card' },
  { id: 4, name: 'Google Ads', category: 'Marketing', date: 'Oct 21, 2024', amount: '-$650.00', status: 'Completed', account: 'Business Card' },
  { id: 5, name: 'Client Invoice #002', category: 'Income', date: 'Oct 20, 2024', amount: '+$3,400.00', status: 'Completed', account: 'Business Checking' },
];

export class InMemoryTransactionRepository implements TransactionRepository {
  async getRecentTransactions(limit: number): Promise<Transaction[]> {
    return MOCK_TRANSACTIONS.slice(0, limit);
  }

  async getAllTransactions(): Promise<Transaction[]> {
    return MOCK_TRANSACTIONS;
  }
}
