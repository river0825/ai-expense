'use client';

import React, { useRef, useCallback, useMemo, useState } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { format } from 'date-fns';
import { Expense } from '@/domain/models/Expense';
import { CurrencyAmount } from './CurrencyAmount';
import { PencilSquareIcon, CheckIcon, XMarkIcon } from '@heroicons/react/24/outline';

interface VirtualExpenseListProps {
  groupedExpenses: Record<string, Expense[]>;
  userHomeCurrency: string;
  onLoadMore: () => void;
  onUpdateExpense: (expense: Expense) => Promise<void>;
  hasMore: boolean;
  isLoading?: boolean;
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

type VirtualItem =
  | { type: 'group-header'; date: string }
  | { type: 'expense-row'; expense: Expense };

export function VirtualExpenseList({
  groupedExpenses,
  userHomeCurrency,
  onLoadMore,
  onUpdateExpense,
  hasMore,
  isLoading = false,
}: VirtualExpenseListProps) {
  const scrollContainerRef = useRef<HTMLDivElement>(null);
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

  // Flatten grouped expenses into virtual items
  const virtualItems = useMemo(() => {
    const items: VirtualItem[] = [];
    Object.entries(groupedExpenses)
      .sort(([dateA], [dateB]) => dateB.localeCompare(dateA))
      .forEach(([dateKey, expenses]) => {
        items.push({ type: 'group-header', date: dateKey });
        expenses.forEach((expense) => {
          items.push({ type: 'expense-row', expense });
        });
      });
    return items;
  }, [groupedExpenses]);

  // Setup virtualizer
  const virtualizer = useVirtualizer({
    count: virtualItems.length,
    getScrollElement: () => scrollContainerRef.current,
    estimateSize: () => 32,
    overscan: 10,
  });

  // Handle scroll event for infinite scroll
  const handleScroll = useCallback(() => {
    if (!scrollContainerRef.current) return;
    const { scrollHeight, scrollTop, clientHeight } = scrollContainerRef.current;
    const distanceFromBottom = scrollHeight - scrollTop - clientHeight;

    if (distanceFromBottom < 200 && hasMore && !isLoading) {
      onLoadMore();
    }
  }, [hasMore, isLoading, onLoadMore]);

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

  // Get all unique currencies and accounts
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

  const virtualItems_array = virtualizer.getVirtualItems();

  return (
    <div
      ref={scrollContainerRef}
      data-testid="virtual-scroller"
      className="w-full h-[600px] overflow-y-auto"
      onScroll={handleScroll}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualItems_array.map((virtualItem) => {
          const item = virtualItems[virtualItem.index];
          if (!item) return null;

          if (item.type === 'group-header') {
            return (
              <div
                key={`header-${item.date}`}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  height: `${virtualItem.size}px`,
                  transform: `translateY(${virtualItem.start}px)`,
                }}
                className="border-b border-border/10 bg-background/30 px-3 py-1 flex items-center"
              >
                <div className="text-xs font-medium text-text/50">
                  {format(new Date(item.date + 'T00:00:00'), 'MMMM dd, yyyy')}
                </div>
              </div>
            );
          }

          const expense = item.expense;
          const isEditing = editingId === expense.id;

          return (
            <div
              key={`expense-${expense.id}`}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: `${virtualItem.size}px`,
                transform: `translateY(${virtualItem.start}px)`,
              }}
              className="border-b border-border/10 hover:bg-background/50 transition-colors flex items-center px-3 gap-3 text-xs"
            >
              {isEditing ? (
                <>
                  <input
                    type="text"
                    value={editForm.description}
                    onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                    className="flex-1 px-1 py-0.5 border border-border/50 rounded bg-background text-xs"
                  />
                  <span className="text-text/70 w-20">{expense.category_name || 'Uncategorized'}</span>
                  <select
                    value={editForm.account}
                    onChange={(e) => setEditForm({ ...editForm, account: e.target.value })}
                    className="w-24 px-1 py-0.5 border border-border/50 rounded bg-background text-xs"
                  >
                    <option value="">{expense.account || 'Select'}</option>
                    {accountOptions.map((acc) => (
                      <option key={acc} value={acc}>
                        {acc}
                      </option>
                    ))}
                  </select>
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
                    className="w-20 px-1 py-0.5 border border-border/50 rounded bg-background text-right text-xs"
                    step="0.01"
                  />
                  <div className="flex gap-1">
                    <button
                      onClick={() => saveEditing(expense)}
                      disabled={isSaving}
                      className="p-0.5 hover:bg-green-500/20 rounded disabled:opacity-50"
                    >
                      <CheckIcon className="w-4 h-4 text-green-600" />
                    </button>
                    <button
                      onClick={cancelEditing}
                      disabled={isSaving}
                      className="p-0.5 hover:bg-red-500/20 rounded disabled:opacity-50"
                    >
                      <XMarkIcon className="w-4 h-4 text-red-600" />
                    </button>
                  </div>
                </>
              ) : (
                <>
                  <span className="flex-1 font-medium text-text truncate">
                    {expense.description}
                  </span>
                  <span className="text-text/70 w-20">{expense.category_name || 'Uncategorized'}</span>
                  <span className="text-text/70 w-24">{expense.account || '-'}</span>
                  <div className="w-20 text-right">
                    <CurrencyAmount
                      amount={expense.home_amount ?? expense.amount}
                      currency={expense.home_currency || userHomeCurrency}
                      originalAmount={expense.original_amount}
                      originalCurrency={expense.original_currency}
                      className="text-xs"
                    />
                  </div>
                  <button
                    onClick={() => startEditing(expense)}
                    className="p-0.5 hover:bg-blue-500/20 rounded"
                  >
                    <PencilSquareIcon className="w-4 h-4 text-blue-600" />
                  </button>
                </>
              )}
            </div>
          );
        })}

        {isLoading && hasMore && (
          <div
            style={{
              position: 'absolute',
              top: `${virtualizer.getTotalSize()}px`,
              left: 0,
              width: '100%',
              height: '32px',
            }}
            className="flex items-center justify-center border-b border-border/10"
          >
            <span className="text-xs text-text/50">Loading more...</span>
          </div>
        )}
      </div>
    </div>
  );
}
