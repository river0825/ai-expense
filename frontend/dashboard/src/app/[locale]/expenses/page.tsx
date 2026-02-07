'use client';

import React, { useEffect, useState, useCallback, useMemo } from 'react';
import { useSearchParams } from 'next/navigation';
import { format } from 'date-fns';
import { DashboardLayout } from '@/components/DashboardLayout';
import { VirtualExpenseList } from '@/components/VirtualExpenseList';
import { AccountFilter } from '@/components/AccountFilter';
import RepositoryFactory from '@/infrastructure/RepositoryFactory';
import { Expense } from '@/domain/models/Expense';
import { getCookie, setCookie } from '@/utils/cookies';
import { ListBulletIcon, CalendarIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';

const BATCH_SIZE = 100;

export default function ExpensesPage() {
  const searchParams = useSearchParams();
  const urlToken = searchParams.get('token');

  // State management
  const [allExpenses, setAllExpenses] = useState<Expense[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [groupBy, setGroupBy] = useState<'date' | 'category'>('date');
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);
  const [selectedAccount, setSelectedAccount] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [offset, setOffset] = useState(0);

  // Token management
  useEffect(() => {
    if (urlToken) {
      setCookie('report_token', urlToken, 604800); // 7 days
      const newUrl = new URL(window.location.href);
      newUrl.searchParams.delete('token');
      window.history.replaceState({}, '', newUrl.toString());
    }
  }, [urlToken]);

  const getToken = useCallback(() => {
    if (urlToken) return urlToken;
    return getCookie('report_token');
  }, [urlToken]);

  // Fetch initial expenses
  useEffect(() => {
    const fetchExpenses = async () => {
      const token = getToken();
      if (!token) {
        setError('Please open the link from your chat to access your expenses.');
        setLoading(false);
        return;
      }

      setLoading(true);
      try {
        const expenseRepo = RepositoryFactory.getExpenseRepository();
        const expenses = await expenseRepo.getExpenses(token);
        setAllExpenses(expenses);
        setHasMore(expenses.length >= BATCH_SIZE);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch expenses:', err);
        setError('Failed to load your expenses. The link may have expired.');
      } finally {
        setLoading(false);
      }
    };

    fetchExpenses();
  }, [getToken]);

  // Handle load more (infinite scroll)
  const handleLoadMore = useCallback(async () => {
    const token = getToken();
    if (!token || isLoadingMore || !hasMore) return;

    setIsLoadingMore(true);
    try {
      const expenseRepo = RepositoryFactory.getExpenseRepository();
      // Note: The current API doesn't support pagination yet, so we'll just mark hasMore as false
      // Once pagination is implemented, this can fetch the next batch
      setHasMore(false);
    } catch (err) {
      console.error('Failed to load more expenses:', err);
    } finally {
      setIsLoadingMore(false);
    }
  }, [getToken, isLoadingMore, hasMore]);

  // Handle expense update
  const handleUpdateExpense = useCallback(
    async (updatedExpense: Expense) => {
      const token = getToken();
      if (!token) return;

      try {
        const expenseRepo = RepositoryFactory.getExpenseRepository();
        await expenseRepo.updateExpense(token, updatedExpense);

        // Update local state
        setAllExpenses((prev) =>
          prev.map((e) => (e.id === updatedExpense.id ? updatedExpense : e))
        );
      } catch (error) {
        console.error('Failed to update expense', error);
        throw error;
      }
    },
    [getToken]
  );

  // Extract unique accounts
  const uniqueAccounts = useMemo(
    () => Array.from(new Set(allExpenses.map((e) => e.account).filter(Boolean) as string[])),
    [allExpenses]
  );

  // Get home currency from expenses or default to USD
  const userHomeCurrency = useMemo(
    () => allExpenses[0]?.home_currency || 'USD',
    [allExpenses]
  );

  // Filter expenses
  const filteredExpenses = useMemo(() => {
    let filtered = allExpenses;

    // Filter by account
    if (selectedAccount) {
      filtered = filtered.filter((e) => e.account === selectedAccount);
    }

    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (e) =>
          e.description.toLowerCase().includes(query) ||
          e.category_name?.toLowerCase().includes(query) ||
          e.account?.toLowerCase().includes(query)
      );
    }

    return filtered;
  }, [allExpenses, selectedAccount, searchQuery]);

  // Group expenses
  const groupedExpenses = useMemo(() => {
    const grouped: Record<string, Expense[]> = {};

    filteredExpenses.forEach((expense) => {
      let key: string;

      if (groupBy === 'date') {
        key = expense.expense_date; // YYYY-MM-DD format
      } else {
        key = expense.category_name || 'Uncategorized';
      }

      if (!grouped[key]) {
        grouped[key] = [];
      }
      grouped[key].push(expense);
    });

    return grouped;
  }, [filteredExpenses, groupBy]);

  // Calculate totals
  const totals = useMemo(() => {
    const total = filteredExpenses.reduce((sum, e) => sum + (e.home_amount ?? e.amount), 0);
    const count = filteredExpenses.length;
    const average = count > 0 ? total / count : 0;

    return {
      total: total.toFixed(2),
      count,
      average: average.toFixed(2),
    };
  }, [filteredExpenses]);

  // Loading state
  if (loading && allExpenses.length === 0) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-screen">
          <div className="flex flex-col items-center gap-3">
            <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary" />
            <p className="text-sm text-text/60">Loading your expenses...</p>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  // Error state
  if (error && allExpenses.length === 0) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-screen p-4">
          <div className="p-6 bg-surface rounded-xl border border-white/10 text-center max-w-md">
            <div className="w-12 h-12 bg-rose-500/10 rounded-full flex items-center justify-center mx-auto mb-4 text-rose-400">
              <CalendarIcon className="w-6 h-6" />
            </div>
            <h2 className="text-xl font-bold mb-2">Access Denied</h2>
            <p className="text-text/60">{error}</p>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="h-screen flex flex-col p-6">
        {/* Header */}
        <div className="mb-4">
          <h1 className="text-2xl font-bold text-text tracking-tight">Expenses</h1>
          <p className="text-text/60 text-sm">
            {filteredExpenses.length} of {allExpenses.length} transactions
          </p>
        </div>

        {/* Sticky Filter */}
        <div className="sticky top-0 z-10 bg-background border-b border-white/10 mb-4 p-3 space-y-3 rounded-lg">
          {/* Search Input - Full Width */}
          <div className="relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-text/40" />
            <input
              type="text"
              placeholder="Search expenses..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 bg-white/5 border border-white/10 rounded-lg text-text placeholder-text/40 focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
            />
          </div>

          {/* Controls Row */}
          <div className="flex flex-wrap gap-2 items-center">
            <span className="text-sm text-text/60 font-medium">Group by:</span>
            {['date', 'category'].map((option) => (
              <button
                key={option}
                onClick={() => setGroupBy(option as 'date' | 'category')}
                className={`px-3 py-1 text-xs font-medium rounded-md transition-all cursor-pointer
                  ${
                    groupBy === option
                      ? 'bg-primary text-white'
                      : 'bg-white/5 text-text/70 hover:bg-white/10 border border-white/10'
                  }
                `}
              >
                {option === 'date' ? 'Date' : 'Category'}
              </button>
            ))}

            <AccountFilter
              accounts={uniqueAccounts}
              selectedAccount={selectedAccount}
              onSelectAccount={setSelectedAccount}
            />
          </div>
        </div>

        {/* Content Area */}
        {filteredExpenses.length === 0 ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <MagnifyingGlassIcon className="w-12 h-12 mx-auto mb-3 opacity-40 text-text/40" />
              <p className="text-text/40">
                {searchQuery || selectedAccount ? 'No expenses match your filters' : 'No expenses found'}
              </p>
            </div>
          </div>
        ) : (
          <>
            <VirtualExpenseList
              groupedExpenses={groupedExpenses}
              userHomeCurrency={userHomeCurrency}
              onLoadMore={handleLoadMore}
              onUpdateExpense={handleUpdateExpense}
              hasMore={hasMore}
              isLoading={isLoadingMore}
            />

            {/* Footer Summary */}
            <div className="mt-4 pt-3 border-t border-white/10 flex items-center justify-between text-sm">
              <span className="text-text/60">
                Showing <span className="font-semibold text-text">{filteredExpenses.length}</span> of{' '}
                <span className="font-semibold text-text">{allExpenses.length}</span> expenses
              </span>
              <span className="font-mono font-bold text-text">
                Total: {new Intl.NumberFormat('en-US', { style: 'currency', currency: userHomeCurrency || 'USD' }).format(
                  filteredExpenses.reduce((sum, e) => sum + (e.home_amount || e.amount), 0)
                )}
              </span>
            </div>
          </>
        )}
      </div>
    </DashboardLayout>
  );
}
