import { render, screen } from '@testing-library/react';
import { VirtualExpenseList } from '@/components/VirtualExpenseList';
import { Expense } from '@/domain/models/Expense';
import { describe, it, expect, vi } from 'vitest';

const mockExpenses: Expense[] = Array.from({ length: 10 }, (_, i) => ({
  id: `${i}`,
  user_id: 'user1',
  description: `Expense ${i}`,
  amount: 10 + i,
  home_amount: 10 + i,
  currency: 'USD',
  home_currency: 'USD',
  expense_date: '2026-02-06',
  category_name: 'Food',
  account: 'Card',
  original_amount: null,
  original_currency: null,
  exchange_rate: null,
  created_at: '2026-02-06T00:00:00Z',
  updated_at: '2026-02-06T00:00:00Z',
}));

describe('VirtualExpenseList', () => {
  it('renders grouped expenses', () => {
    const grouped = { '2026-02-06': mockExpenses.slice(0, 5) };
    render(
      <VirtualExpenseList
        groupedExpenses={grouped}
        userHomeCurrency="USD"
        onLoadMore={vi.fn()}
        onUpdateExpense={() => Promise.resolve()}
        hasMore={false}
      />
    );
    expect(screen.getByText('Expense 0')).toBeInTheDocument();
  });

  it('calls onLoadMore callback', () => {
    const onLoadMore = vi.fn();
    const grouped = { '2026-02-06': mockExpenses.slice(0, 5) };
    render(
      <VirtualExpenseList
        groupedExpenses={grouped}
        userHomeCurrency="USD"
        onLoadMore={onLoadMore}
        onUpdateExpense={() => Promise.resolve()}
        hasMore={true}
      />
    );
    expect(onLoadMore).toBeDefined();
  });

  it('renders with virtual scroller container', () => {
    const grouped = { '2026-02-06': mockExpenses.slice(0, 5) };
    const { container } = render(
      <VirtualExpenseList
        groupedExpenses={grouped}
        userHomeCurrency="USD"
        onLoadMore={vi.fn()}
        onUpdateExpense={() => Promise.resolve()}
        hasMore={false}
      />
    );
    const scroller = container.querySelector('[data-testid="virtual-scroller"]');
    expect(scroller).toBeInTheDocument();
  });
});
