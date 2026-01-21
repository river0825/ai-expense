export interface DashboardStats {
  totalBalance: {
    value: string;
    change: string;
    isPositive: boolean;
  };
  totalIncome: {
    value: string;
    change: string;
    isPositive: boolean;
  };
  totalExpenses: {
    value: string;
    change: string;
    isPositive: boolean;
  };
}
