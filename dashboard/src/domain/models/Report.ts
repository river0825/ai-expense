export interface CategoryBreakdown {
  category: string;
  total: number;
  count: number;
  percentage: number;
}

export interface DailyBreakdown {
  date: string;
  total: number;
  count: number;
  amount: number;
}

export interface ExpenseDetail {
  id: string;
  description: string;
  amount: number;
  category: string;
  date: string;
}

export interface ExpenseReport {
  user_id: string;
  report_type: string;
  period: string;
  start_date: string;
  end_date: string;
  total_expenses: number;
  transaction_count: number;
  average_expense: number;
  highest_expense: number;
  lowest_expense: number;
  category_breakdown: CategoryBreakdown[];
  daily_breakdown: DailyBreakdown[];
  top_expenses: ExpenseDetail[];
  generated_at: string;
}
