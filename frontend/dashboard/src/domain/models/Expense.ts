export interface Expense {
  id: string;
  user_id: string;
  description: string;
  amount: number;
  category_id?: string;
  category_name?: string;
  account?: string;
  expense_date: string;  // ISO date string
  created_at?: string;
  currency?: string;
  original_amount?: number;
  home_amount?: number;
  home_currency?: string;
  exchange_rate?: number;
}

export interface CategoryTotal {
  category_id?: string;
  category_name: string;
  total: number;
  count: number;
  percentage: number;
}

export interface TrendDataPoint {
  date: string;
  amount: number;
  count: number;
}

// Date range preset type
export type DatePreset = 'today' | 'week' | 'month' | 'last7' | 'last30' | 'last90' | 'custom';
