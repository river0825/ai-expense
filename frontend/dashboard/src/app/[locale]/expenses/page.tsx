'use client';

import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useSearchParams } from 'next/navigation';
import { DashboardLayout } from '@/components/DashboardLayout';
import { ExpenseList } from '@/components/ExpenseList';
import RepositoryFactory from '@/infrastructure/RepositoryFactory';
import { Expense } from '@/domain/models/Expense';
import { getCookie, setCookie } from '@/utils/cookies';

export default function ExpensesPage() {
  const searchParams = useSearchParams();
  const urlToken = searchParams.get('token');
  const editExpenseId = searchParams.get('edit'); // Deep link from LINE flex message

  // Capture token on first render so URL cleanup doesn't cause re-fetches
  const tokenRef = useRef<string | null>(urlToken || getCookie('report_token'));

  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Token management + URL cleanup (runs once)
  useEffect(() => {
    if (urlToken) {
      setCookie('report_token', urlToken, 604800); // 7 days
      tokenRef.current = urlToken;
    }

    // Clean up query params from URL
    const newUrl = new URL(window.location.href);
    let changed = false;
    if (newUrl.searchParams.has('token')) { newUrl.searchParams.delete('token'); changed = true; }
    if (newUrl.searchParams.has('edit')) { newUrl.searchParams.delete('edit'); changed = true; }
    if (changed) window.history.replaceState({}, '', newUrl.toString());
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const getToken = useCallback(() => {
    return tokenRef.current;
  }, []);

  // Fetch expenses (runs once)
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
        setExpenses((prev) =>
          prev.map((e) => (e.id === updatedExpense.id ? updatedExpense : e))
        );
      } catch (error) {
        console.error('Failed to update expense', error);
        throw error;
      }
    },
    [getToken]
  );

  // Handle expense delete
  const handleDeleteExpense = useCallback(
    async (expense: Expense) => {
      const token = getToken();
      if (!token) return;

      try {
        const expenseRepo = RepositoryFactory.getExpenseRepository();
        await expenseRepo.deleteExpense(token, expense.id, expense.user_id);

        // Remove from local state
        setExpenses((prev) => prev.filter((e) => e.id !== expense.id));
      } catch (error) {
        console.error('Failed to delete expense', error);
        throw error;
      }
    },
    [getToken]
  );

  // Loading state
  if (loading && expenses.length === 0) {
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
  if (error && expenses.length === 0) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-screen p-4">
          <div className="p-6 bg-surface rounded-xl border border-white/10 text-center max-w-md">
            <h2 className="text-xl font-bold mb-2">Access Denied</h2>
            <p className="text-text/60">{error}</p>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="h-[calc(100vh-5rem)] flex flex-col p-3 sm:p-6">
        {/* Header - hidden on mobile to save space */}
        <div className="mb-4 hidden sm:block">
          <h1 className="text-2xl font-bold text-text tracking-tight">Expenses</h1>
          <p className="text-text/60 text-sm">{expenses.length} transactions</p>
        </div>

        {/* Expense List Component - Handles search, filtering, grouping, editing all internally */}
        <ExpenseList
          expenses={expenses}
          onUpdateExpense={handleUpdateExpense}
          onDeleteExpense={handleDeleteExpense}
          className="flex-1"
          initialEditingId={editExpenseId}
        />
      </div>
    </DashboardLayout>
  );
}
