'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useSearchParams } from 'next/navigation';
import { DashboardLayout } from '@/components/DashboardLayout';
import { ExpenseList } from '@/components/ExpenseList';
import RepositoryFactory from '@/infrastructure/RepositoryFactory';
import { Expense } from '@/domain/models/Expense';
import { getCookie, setCookie } from '@/utils/cookies';

export default function ExpensesPage() {
  const searchParams = useSearchParams();
  const urlToken = searchParams.get('token');

  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  // Fetch expenses
  useEffect(() => {
    const fetchExpenses = async () => {
      const token = getToken();
      if (!token) {
        setError('Please open the link from your chat to access your expenses.');
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        const expenseRepo = RepositoryFactory.getExpenseRepository();
        const data = await expenseRepo.getExpenses(token);
        setExpenses(data || []);
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
          <p className="text-text/60 text-sm">{expenses.length} transactions</p>
        </div>

        {/* Expense List Component - Handles search, filtering, grouping, editing all internally */}
        <ExpenseList
          expenses={expenses}
          onUpdateExpense={handleUpdateExpense}
          className="flex-1"
        />
      </div>
    </DashboardLayout>
  );
}
