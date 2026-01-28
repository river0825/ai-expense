import axios from 'axios';
import { format, startOfDay, endOfDay, eachDayOfInterval, eachWeekOfInterval, eachMonthOfInterval } from 'date-fns';
import { ExpenseRepository } from '@/domain/repositories/ExpenseRepository';
import { Expense, CategoryTotal, TrendDataPoint } from '@/domain/models/Expense';
import { ExpenseReport, ExpenseDetail } from '@/domain/models/Report';

export class HttpExpenseRepository implements ExpenseRepository {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://aiexpense-996531141309.us-central1.run.app';
  }


  static mapToExpenses(report: ExpenseReport, categoryId?: string): Expense[] {
      // Convert ExpenseDetail[] to Expense[]
      let expenses: Expense[] = report.top_expenses.map((detail: ExpenseDetail) => ({
        id: detail.id,
        user_id: report.user_id,
        description: detail.description,
        amount: detail.amount,
        category_name: detail.category,
        expense_date: detail.date,
        account: detail.account,
      }));

      // Filter by category if specified
      if (categoryId) {
        expenses = expenses.filter(e => e.category_id === categoryId);
      }
      return expenses;
  }

  static mapToCategoryTotals(report: ExpenseReport): CategoryTotal[] {
      // Convert CategoryBreakdown[] to CategoryTotal[]
      return report.category_breakdown.map(cb => ({
        category_name: cb.category,
        total: cb.total,
        count: cb.count,
        percentage: cb.percentage,
      }));
  }

  static mapToTrendData(report: ExpenseReport, groupBy: 'day' | 'week' | 'month', startDate: Date, endDate: Date): TrendDataPoint[] {
      // Use daily_breakdown from backend
      const dailyData: TrendDataPoint[] = report.daily_breakdown.map(db => ({
        date: format(new Date(db.date), 'yyyy-MM-dd'),
        amount: db.total,
        count: db.count,
      }));

      // Aggregate by week or month if needed
      if (groupBy === 'day') {
        return dailyData;
      } else if (groupBy === 'week') {
        // We need to access the private helper method or move it to static
        return HttpExpenseRepository.aggregateByWeek(dailyData, startDate, endDate);
      } else {
        return HttpExpenseRepository.aggregateByMonth(dailyData, startDate, endDate);
      }
  }

  async getExpenses(
    token: string,
    startDate?: Date,
    endDate?: Date,
    categoryId?: string
  ): Promise<Expense[]> {
    try {
      // Use the /api/reports/summary endpoint which returns all expenses (after backend fix)
      let url = `${this.baseURL}/api/reports/summary?token=${token}`;
      
      if (startDate) {
        url += `&start_date=${format(startDate, 'yyyy-MM-dd')}`;
      }
      if (endDate) {
        url += `&end_date=${format(endDate, 'yyyy-MM-dd')}`;
      }

      const response = await axios.get<{ status: string; data: ExpenseReport }>(url);
      const report = response.data.data;
      
      return HttpExpenseRepository.mapToExpenses(report, categoryId);
    } catch (error) {
      console.error('Failed to fetch expenses:', error);
      throw error;
    }
  }

  async getCategoryTotals(
    token: string,
    startDate?: Date,
    endDate?: Date
  ): Promise<CategoryTotal[]> {
    try {
      let url = `${this.baseURL}/api/reports/summary?token=${token}`;
      
      if (startDate) {
        url += `&start_date=${format(startDate, 'yyyy-MM-dd')}`;
      }
      if (endDate) {
        url += `&end_date=${format(endDate, 'yyyy-MM-dd')}`;
      }

      const response = await axios.get<{ status: string; data: ExpenseReport }>(url);
      const report = response.data.data;
      
      return HttpExpenseRepository.mapToCategoryTotals(report);
    } catch (error) {
      console.error('Failed to fetch category totals:', error);
      throw error;
    }
  }

  async getTrendData(
    token: string,
    startDate: Date,
    endDate: Date,
    groupBy: 'day' | 'week' | 'month'
  ): Promise<TrendDataPoint[]> {
    try {
      let url = `${this.baseURL}/api/reports/summary?token=${token}`;
      url += `&start_date=${format(startDate, 'yyyy-MM-dd')}`;
      url += `&end_date=${format(endDate, 'yyyy-MM-dd')}`;

      const response = await axios.get<{ status: string; data: ExpenseReport }>(url);
      const report = response.data.data;
      
      return HttpExpenseRepository.mapToTrendData(report, groupBy, startDate, endDate);
    } catch (error) {
      console.error('Failed to fetch trend data:', error);
      throw error;
    }
  }

  private static aggregateByWeek(dailyData: TrendDataPoint[], startDate: Date, endDate: Date): TrendDataPoint[] {
    const weeks = eachWeekOfInterval({ start: startDate, end: endDate }, { weekStartsOn: 1 });
    const weekMap = new Map<string, { amount: number; count: number }>();

    weeks.forEach(week => {
      const weekKey = format(week, 'yyyy-MM-dd');
      weekMap.set(weekKey, { amount: 0, count: 0 });
    });

    dailyData.forEach(day => {
      const dayDate = new Date(day.date);
      const weekStart = weeks.find(w => {
        const weekEnd = new Date(w);
        weekEnd.setDate(weekEnd.getDate() + 6);
        return dayDate >= w && dayDate <= weekEnd;
      });

      if (weekStart) {
        const weekKey = format(weekStart, 'yyyy-MM-dd');
        const existing = weekMap.get(weekKey)!;
        weekMap.set(weekKey, {
          amount: existing.amount + day.amount,
          count: existing.count + day.count,
        });
      }
    });

    return Array.from(weekMap.entries()).map(([date, data]) => ({
      date,
      amount: data.amount,
      count: data.count,
    }));
  }

  private static aggregateByMonth(dailyData: TrendDataPoint[], startDate: Date, endDate: Date): TrendDataPoint[] {
    const months = eachMonthOfInterval({ start: startDate, end: endDate });
    const monthMap = new Map<string, { amount: number; count: number }>();

    months.forEach(month => {
      const monthKey = format(month, 'yyyy-MM');
      monthMap.set(monthKey, { amount: 0, count: 0 });
    });

    dailyData.forEach(day => {
      const monthKey = day.date.substring(0, 7); // yyyy-MM
      const existing = monthMap.get(monthKey);
      if (existing) {
        monthMap.set(monthKey, {
          amount: existing.amount + day.amount,
          count: existing.count + day.count,
        });
      }
    });

    return Array.from(monthMap.entries()).map(([date, data]) => ({
      date: date + '-01', // First day of month
      amount: data.amount,
      count: data.count,
    }));
  }

  async updateExpense(token: string, expense: Expense): Promise<void> {
    try {
      // Backend expects ID in the body for PUT /api/expenses, not as a path parameter
      const url = `${this.baseURL}/api/expenses?token=${token}`;
      
      const payload: any = {
        id: expense.id,
        user_id: expense.user_id,
        description: expense.description,
        amount: expense.amount,
      };

      // Only include account if it exists
      if (expense.account) {
        payload.account = expense.account;
      }

      await axios.put(url, payload);
    } catch (error) {
      console.error('Failed to update expense:', error);
      throw error;
    }
  }
}
