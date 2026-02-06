import { render, screen } from '@testing-library/react';
import { ExpenseListTable } from '@/components/ExpenseListTable';
import { Expense } from '@/domain/models/Expense';
import { describe, it, expect } from 'vitest';

const mockExpense: Expense = {
  id: '1',
  user_id: 'user1',
  description: 'Coffee',
  amount: 5.5,
  home_amount: 5.5,
  currency: 'USD',
  home_currency: 'USD',
  expense_date: '2026-02-06',
  category_name: 'Food',
  account: 'Credit Card',
  original_amount: undefined,
  original_currency: undefined,
  exchange_rate: undefined,
  created_at: '2026-02-06T00:00:00Z',
  updated_at: '2026-02-06T00:00:00Z',
};

describe('ExpenseListTable', () => {
  it('renders table header with columns', () => {
    render(
      <ExpenseListTable
        groupedExpenses={{ '2026-02-06': [mockExpense] }}
        userHomeCurrency="USD"
        onUpdateExpense={() => Promise.resolve()}
      />
    );
    expect(screen.getByText('Description')).toBeInTheDocument();
    expect(screen.getByText('Category')).toBeInTheDocument();
    expect(screen.getByText('Account')).toBeInTheDocument();
    expect(screen.getByText('Date')).toBeInTheDocument();
    expect(screen.getByText('Amount')).toBeInTheDocument();
  });

  it('renders expense row with data', () => {
    render(
      <ExpenseListTable
        groupedExpenses={{ '2026-02-06': [mockExpense] }}
        userHomeCurrency="USD"
        onUpdateExpense={() => Promise.resolve()}
      />
    );
    expect(screen.getByText('Coffee')).toBeInTheDocument();
    expect(screen.getByText('Food')).toBeInTheDocument();
    expect(screen.getByText('Credit Card')).toBeInTheDocument();
  });

  it('renders group headers with date when grouped by date', () => {
    render(
      <ExpenseListTable
        groupedExpenses={{ '2026-02-06': [mockExpense] }}
        userHomeCurrency="USD"
        onUpdateExpense={() => Promise.resolve()}
      />
    );
    expect(screen.getByText(/February 06, 2026/)).toBeInTheDocument();
  });
});
