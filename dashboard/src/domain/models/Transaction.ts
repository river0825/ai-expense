export interface Transaction {
  id: number;
  name: string;
  category: string;
  date: string;
  amount: string; // Keeping as string for now to match current mock data format (e.g. "-$240.00")
  status: 'Completed' | 'Pending';
  account: string;
}
