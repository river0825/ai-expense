import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { StickyExpenseFilter } from '@/components/StickyExpenseFilter';
import { describe, it, expect, vi } from 'vitest';

describe('StickyExpenseFilter', () => {
  it('renders search input', () => {
    render(
      <StickyExpenseFilter
        searchQuery=""
        onSearchChange={() => {}}
        groupBy="date"
        onGroupByChange={() => {}}
        selectedDate={new Date()}
        onDateSelect={() => {}}
        selectedAccount={null}
        onAccountSelect={() => {}}
        accounts={[]}
      />
    );
    expect(screen.getByPlaceholderText('Search expenses...')).toBeInTheDocument();
  });

  it('renders group by buttons', () => {
    render(
      <StickyExpenseFilter
        searchQuery=""
        onSearchChange={() => {}}
        groupBy="date"
        onGroupByChange={() => {}}
        selectedDate={new Date()}
        onDateSelect={() => {}}
        selectedAccount={null}
        onAccountSelect={() => {}}
        accounts={[]}
      />
    );
    expect(screen.getByText('Date')).toBeInTheDocument();
    expect(screen.getByText('Category')).toBeInTheDocument();
  });

  it('calls onGroupByChange when group button clicked', async () => {
    const onGroupByChange = vi.fn();
    const user = userEvent.setup();

    render(
      <StickyExpenseFilter
        searchQuery=""
        onSearchChange={() => {}}
        groupBy="date"
        onGroupByChange={onGroupByChange}
        selectedDate={new Date()}
        onDateSelect={() => {}}
        selectedAccount={null}
        onAccountSelect={() => {}}
        accounts={[]}
      />
    );

    await user.click(screen.getByText('Category'));
    expect(onGroupByChange).toHaveBeenCalledWith('category');
  });

  it('is sticky positioned', () => {
    const { container } = render(
      <StickyExpenseFilter
        searchQuery=""
        onSearchChange={() => {}}
        groupBy="date"
        onGroupByChange={() => {}}
        selectedDate={new Date()}
        onDateSelect={() => {}}
        selectedAccount={null}
        onAccountSelect={() => {}}
        accounts={[]}
      />
    );
    const filterBox = container.querySelector('[data-testid="sticky-filter"]');
    expect(filterBox).toHaveClass('sticky', 'top-0', 'z-10');
  });
});
