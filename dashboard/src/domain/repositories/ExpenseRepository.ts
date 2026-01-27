import { Expense, CategoryTotal, TrendDataPoint } from '../models/Expense';

export interface ExpenseRepository {
  /**
   * Get all expenses for a user within a date range
   */
  getExpenses(
    token: string,
    startDate?: Date,
    endDate?: Date,
    categoryId?: string
  ): Promise<Expense[]>;
  
  /**
   * Get category totals aggregated from expenses
   */
  getCategoryTotals(
    token: string,
    startDate?: Date,
    endDate?: Date
  ): Promise<CategoryTotal[]>;
  
  /**
   * Get trend data grouped by time period
   */
  getTrendData(
    token: string,
    startDate: Date,
    endDate: Date,
    groupBy: 'day' | 'week' | 'month'
  ): Promise<TrendDataPoint[]>;
}
