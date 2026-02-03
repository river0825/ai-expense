import { Transaction } from '../models/Transaction';

export interface TransactionRepository {
  getRecentTransactions(limit: number): Promise<Transaction[]>;
  getAllTransactions(): Promise<Transaction[]>;
}
