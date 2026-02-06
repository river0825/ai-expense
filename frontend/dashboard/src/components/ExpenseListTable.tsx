'use client';

import React, { useState, useMemo } from 'react';
import { format } from 'date-fns';
import { Expense } from '@/domain/models/Expense';
import { CurrencyAmount } from './CurrencyAmount';
import { PencilSquareIcon, CheckIcon, XMarkIcon } from '@heroicons/react/24/outline';

interface ExpenseListTableProps {
  groupedExpenses: Record<string, Expense[]>;
  userHomeCurrency: string;
  onUpdateExpense: (expense: Expense) => Promise<void>;
}

type EditFormState = {
  description: string;
  originalAmount: number;
  currency: string;
  account: string;
  conversionRate: number;
  homePreview: number;
  homeCurrency: string;
};

export function ExpenseListTable({
  groupedExpenses,
  userHomeCurrency,
  onUpdateExpense,
}: ExpenseListTableProps) {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<EditFormState>({
    description: '',
    originalAmount: 0,
    currency: userHomeCurrency,
    account: '',
    conversionRate: 1,
    homePreview: 0,
    homeCurrency: userHomeCurrency,
  });
  const [isSaving, setIsSaving] = useState(false);

  const determineInitialRate = (expense: Expense, initialCurrency: string, homeCurrency: string) => {
    if (initialCurrency === homeCurrency) {
      return 1;
    }
    if (expense.exchange_rate && expense.exchange_rate > 0) {
      return expense.exchange_rate;
    }
    const original = expense.original_amount ?? expense.amount;
    const home = expense.home_amount ?? expense.amount;
    if (original && original !== 0) {
      return home / original;
    }
    return 1;
  };

  const startEditing = (expense: Expense) => {
    setEditingId(expense.id);
    const homeCurrency = expense.home_currency || userHomeCurrency;
    const initialOriginalAmount = expense.original_amount ?? expense.amount;
    const initialCurrency = expense.original_currency || expense.currency || homeCurrency;
    const conversionRate = determineInitialRate(expense, initialCurrency, homeCurrency);
    const preview = (initialOriginalAmount || 0) * conversionRate;

    setEditForm({
      description: expense.description,
      originalAmount: initialOriginalAmount,
      currency: initialCurrency,
      account: expense.account || '',
      conversionRate,
      homePreview: Number.isNaN(preview) ? 0 : preview,
      homeCurrency,
    });
  };

  const cancelEditing = () => {
    setEditingId(null);
  };

  const saveEditing = async (originalExpense: Expense) => {
    if (!onUpdateExpense) return;

    try {
      setIsSaving(true);
      const originalAmountValue = editForm.originalAmount || 0;
      const conversionRate = editForm.conversionRate || 1;
      const derivedHomeAmount = originalAmountValue * conversionRate;
      const updatedExpense: Expense = {
        ...originalExpense,
        description: editForm.description,
        original_amount: originalAmountValue,
        original_currency: editForm.currency,
        currency: editForm.currency,
        home_amount: derivedHomeAmount,
        home_currency: editForm.homeCurrency || originalExpense.home_currency || userHomeCurrency,
        amount: derivedHomeAmount,
        account: editForm.account,
      };
      await onUpdateExpense(updatedExpense);
      setEditingId(null);
    } catch (error) {
      console.error('Failed to update expense', error);
    } finally {
      setIsSaving(false);
    }
  };

  // Get all unique currencies from expenses
  const currencyOptions = useMemo(() => {
    const unique = new Set<string>();
    Object.values(groupedExpenses).forEach((expenses) => {
      expenses.forEach((expense) => {
        const currency = expense.original_currency || expense.currency || expense.home_currency;
        if (currency) {
          unique.add(currency);
        }
      });
    });
    unique.add(userHomeCurrency);
    return Array.from(unique);
  }, [groupedExpenses, userHomeCurrency]);

  const accountOptions = useMemo(() => {
    const unique = new Set<string>();
    Object.values(groupedExpenses).forEach((expenses) => {
      expenses.forEach((expense) => {
        if (expense.account) {
          unique.add(expense.account);
        }
      });
    });
    return Array.from(unique);
  }, [groupedExpenses]);

  return (
    <div className="w-full overflow-x-auto">
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-border/20 bg-background/50">
            <th className="text-left text-xs font-semibold text-text/70 px-3 py-2 h-8">Description</th>
            <th className="text-left text-xs font-semibold text-text/70 px-3 py-2 h-8">Category</th>
            <th className="text-left text-xs font-semibold text-text/70 px-3 py-2 h-8">Account</th>
            <th className="text-left text-xs font-semibold text-text/70 px-3 py-2 h-8">Date</th>
            <th className="text-right text-xs font-semibold text-text/70 px-3 py-2 h-8">Amount</th>
            <th className="text-center text-xs font-semibold text-text/70 px-3 py-2 h-8 w-12">Action</th>
          </tr>
        </thead>
        <tbody>
          {Object.entries(groupedExpenses).map(([dateKey, expenses]) => (
            <React.Fragment key={dateKey}>
              {/* Group Header */}
              <tr className="border-b border-border/10 bg-background/30">
                <td colSpan={6} className="px-3 py-1 text-xs font-medium text-text/50">
                  {format(new Date(dateKey + 'T00:00:00'), 'MMMM dd, yyyy')}
                </td>
              </tr>

              {/* Expense Rows */}
              {expenses.map((expense) => {
                const isEditing = editingId === expense.id;
                return (
                  <tr
                    key={expense.id}
                    className="border-b border-border/10 h-8 hover:bg-background/50 transition-colors"
                  >
                    {isEditing ? (
                      <>
                        {/* Edit Mode */}
                        <td className="px-3 py-1">
                          <input
                            type="text"
                            value={editForm.description}
                            onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                            className="w-full px-1 py-0.5 text-xs border border-border/50 rounded bg-background"
                          />
                        </td>
                        <td className="px-3 py-1 text-xs text-text/70">{expense.category_name || 'Uncategorized'}</td>
                        <td className="px-3 py-1">
                          <select
                            value={editForm.account}
                            onChange={(e) => setEditForm({ ...editForm, account: e.target.value })}
                            className="w-full px-1 py-0.5 text-xs border border-border/50 rounded bg-background"
                          >
                            <option value="">{expense.account || 'Select account'}</option>
                            {accountOptions.map((acc) => (
                              <option key={acc} value={acc}>
                                {acc}
                              </option>
                            ))}
                          </select>
                        </td>
                        <td className="px-3 py-1 text-xs text-text/70">{dateKey}</td>
                        <td className="px-3 py-1">
                          <input
                            type="number"
                            value={editForm.originalAmount}
                            onChange={(e) => {
                              const amount = parseFloat(e.target.value) || 0;
                              const preview = amount * editForm.conversionRate;
                              setEditForm({
                                ...editForm,
                                originalAmount: amount,
                                homePreview: Number.isNaN(preview) ? 0 : preview,
                              });
                            }}
                            className="w-full px-1 py-0.5 text-xs border border-border/50 rounded bg-background text-right"
                            step="0.01"
                          />
                        </td>
                        <td className="px-3 py-1 flex gap-1 justify-center h-8">
                          <button
                            onClick={() => saveEditing(expense)}
                            disabled={isSaving}
                            className="p-0.5 hover:bg-green-500/20 rounded disabled:opacity-50"
                            title="Save"
                          >
                            <CheckIcon className="w-4 h-4 text-green-600" />
                          </button>
                          <button
                            onClick={cancelEditing}
                            disabled={isSaving}
                            className="p-0.5 hover:bg-red-500/20 rounded disabled:opacity-50"
                            title="Cancel"
                          >
                            <XMarkIcon className="w-4 h-4 text-red-600" />
                          </button>
                        </td>
                      </>
                    ) : (
                      <>
                        {/* View Mode */}
                        <td className="px-3 py-1 text-xs text-text font-medium truncate">
                          {expense.description}
                        </td>
                        <td className="px-3 py-1 text-xs text-text/70">{expense.category_name || 'Uncategorized'}</td>
                        <td className="px-3 py-1 text-xs text-text/70">{expense.account || '-'}</td>
                        <td className="px-3 py-1 text-xs text-text/70">{dateKey}</td>
                        <td className="px-3 py-1 text-right">
                          <CurrencyAmount
                            amount={expense.home_amount ?? expense.amount}
                            currency={expense.home_currency || userHomeCurrency}
                            originalAmount={expense.original_amount}
                            originalCurrency={expense.original_currency}
                            className="text-xs"
                          />
                        </td>
                        <td className="px-3 py-1 text-center h-8 flex items-center justify-center">
                          <button
                            onClick={() => startEditing(expense)}
                            className="p-0.5 hover:bg-blue-500/20 rounded opacity-0 group-hover:opacity-100 transition-opacity"
                            title="Edit"
                          >
                            <PencilSquareIcon className="w-4 h-4 text-blue-600" />
                          </button>
                        </td>
                      </>
                    )}
                  </tr>
                );
              })}
            </React.Fragment>
          ))}
        </tbody>
      </table>
    </div>
  );
}
