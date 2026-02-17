'use client';

import React, { useState, useMemo, useEffect, useRef } from 'react';
import { format } from 'date-fns';
import { Expense } from '@/domain/models/Expense';
import { CurrencyAmount } from './CurrencyAmount';
import {
  MagnifyingGlassIcon,
  FunnelIcon,
  ChevronUpIcon,
  ChevronDownIcon,
  TagIcon,
  CalendarIcon,
  CurrencyDollarIcon,
  PencilSquareIcon,
  CheckIcon,
  XMarkIcon,
  CreditCardIcon,
  BanknotesIcon,
  WalletIcon,
  GlobeAltIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';

interface ExpenseListProps {
  expenses: Expense[];
  onCategoryFilter?: (categoryName: string | null) => void;
  onUpdateExpense?: (expense: Expense) => Promise<void>;
  onDeleteExpense?: (expense: Expense) => Promise<void>;
  className?: string;
  initialEditingId?: string | null;
}

type SortField = 'date' | 'amount' | 'category';
type SortDirection = 'asc' | 'desc';

export function ExpenseList({ expenses, onCategoryFilter, onUpdateExpense, onDeleteExpense, className = '', initialEditingId = null }: ExpenseListProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [sortField, setSortField] = useState<SortField>('date');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [groupBy, setGroupBy] = useState<'none' | 'category' | 'date'>('date');

  // Editing state
  type EditFormState = {
    description: string;
    originalAmount: number;
    currency: string;
    account: string;
    conversionRate: number;
    homePreview: number;
    homeCurrency: string;
  };

  const createEmptyEditForm = (): EditFormState => ({
    description: '',
    originalAmount: 0,
    currency: 'TWD',
    account: '',
    conversionRate: 1,
    homePreview: 0,
    homeCurrency: 'TWD',
  });

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<EditFormState>(createEmptyEditForm);
  const [isSaving, setIsSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const deepLinkHandled = useRef(false);

  const userHomeCurrency = useMemo(() => {
    for (const expense of expenses) {
      if (expense.home_currency) {
        return expense.home_currency;
      }
    }
    return 'TWD';
  }, [expenses]);

  const currencyOptions = useMemo(() => {
    const unique = new Set<string>();
    expenses.forEach((expense) => {
      const currency = expense.original_currency || expense.currency || expense.home_currency;
      if (currency) {
        unique.add(currency);
      }
    });
    unique.add(userHomeCurrency);
    return Array.from(unique);
  }, [expenses, userHomeCurrency]);

  const formatHomePreview = (value: number, currency: string) => {
    if (!value || Number.isNaN(value)) {
      return `≈ ${currency} 0.00`;
    }
    try {
      const locale = currency === 'TWD' ? 'zh-TW' : 'en-US';
      return new Intl.NumberFormat(locale, {
        style: 'currency',
        currency,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(value);
    } catch {
      return `${currency} ${value.toFixed(2)}`;
    }
  };

  const formatCurrencyAmount = (value: number, currency: string) => {
    try {
      const locale = currency === 'TWD' ? 'zh-TW' : 'en-US';
      return new Intl.NumberFormat(locale, {
        style: 'currency',
        currency,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(value);
    } catch {
      if (Number.isNaN(value)) {
        return `${currency} 0.00`;
      }
      return `${currency} ${value.toFixed(2)}`;
    }
  };

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

  const accountOptions = useMemo(() => {
    const unique = new Set<string>();
    expenses.forEach((expense) => {
      if (expense.account) {
        unique.add(expense.account);
      }
    });
    return Array.from(unique);
  }, [expenses]);

  // Auto-trigger editing when initialEditingId is provided (deep link from LINE flex message)
  useEffect(() => {
    if (deepLinkHandled.current || !initialEditingId || expenses.length === 0) return;
    const expense = expenses.find(e => e.id === initialEditingId);
    if (!expense) return;

    deepLinkHandled.current = true;
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
  }, [initialEditingId, expenses, userHomeCurrency]);

  const startEditing = (expense: Expense, e: React.MouseEvent) => {
    e.stopPropagation();
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

  const cancelEditing = (e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    setEditingId(null);
    setEditForm(createEmptyEditForm());
  };

  const saveEditing = async (originalExpense: Expense, e: React.MouseEvent) => {
    e.stopPropagation();
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

  const handleDelete = async (expense: Expense, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!onDeleteExpense) return;
    if (!confirm(`Delete "${expense.description}"?`)) return;

    try {
      setDeletingId(expense.id);
      await onDeleteExpense(expense);
    } catch (error) {
      console.error('Failed to delete expense', error);
    } finally {
      setDeletingId(null);
    }
  };

  // Filter and sort expenses
  const processedExpenses = useMemo(() => {
    let filtered = expenses.filter(expense =>
      expense.description.toLowerCase().includes(searchQuery.toLowerCase())
    );

    // Sort
    filtered.sort((a, b) => {
      let comparison = 0;
      if (sortField === 'date') {
        comparison = new Date(a.expense_date).getTime() - new Date(b.expense_date).getTime();
      } else if (sortField === 'amount') {
        comparison = a.amount - b.amount;
      } else if (sortField === 'category') {
        comparison = (a.category_name || 'Uncategorized').localeCompare(b.category_name || 'Uncategorized');
      }
      return sortDirection === 'asc' ? comparison : -comparison;
    });

    return filtered;
  }, [expenses, searchQuery, sortField, sortDirection]);

  // Group expenses
  const groupedExpenses = useMemo(() => {
    if (groupBy === 'none') {
      return { 'All Expenses': processedExpenses };
    } else if (groupBy === 'category') {
      const groups: Record<string, Expense[]> = {};
      processedExpenses.forEach(expense => {
        const category = expense.category_name || 'Uncategorized';
        if (!groups[category]) groups[category] = [];
        groups[category].push(expense);
      });
      return groups;
    } else {
      // Group by date
      const groups: Record<string, Expense[]> = {};
      processedExpenses.forEach(expense => {
        const dateKey = format(new Date(expense.expense_date), 'yyyy-MM-dd');
        if (!groups[dateKey]) groups[dateKey] = [];
        groups[dateKey].push(expense);
      });
      return groups;
    }
  }, [processedExpenses, groupBy]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return null;
    return sortDirection === 'asc' ? (
      <ChevronUpIcon className="w-4 h-4" />
    ) : (
      <ChevronDownIcon className="w-4 h-4" />
    );
  };

  return (
    <div className={`flex flex-col h-full ${className}`}>
      {/* Controls - hidden on mobile */}
      <div className="mb-4 space-y-3 hidden sm:block">
        {/* Search */}
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

        {/* Filters and Grouping */}
        <div className="flex flex-wrap gap-2 items-center">
          <span className="text-sm text-text/60 font-medium">Group by:</span>
          {['none', 'category', 'date'].map((option) => (
            <button
              key={option}
              onClick={() => setGroupBy(option as typeof groupBy)}
              className={`px-3 py-1 text-xs font-medium rounded-md transition-all cursor-pointer
                ${groupBy === option
                  ? 'bg-primary text-white'
                  : 'bg-white/5 text-text/70 hover:bg-white/10 border border-white/10'
                }
              `}
            >
              {option === 'none' ? 'None' : option.charAt(0).toUpperCase() + option.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {/* Expense List */}
      <div className="flex-1 overflow-y-auto space-y-4 pr-2 custom-scrollbar">
        {Object.keys(groupedExpenses).length === 0 ? (
          <div className="text-center py-12 text-text/40">
            <MagnifyingGlassIcon className="w-12 h-12 mx-auto mb-3 opacity-40" />
            <p>No expenses found</p>
          </div>
        ) : (
          Object.entries(groupedExpenses).map(([groupName, groupExpenses]) => {
            const groupTotal = groupExpenses.reduce((sum, expense) => {
              return sum + (expense.home_amount ?? expense.amount);
            }, 0);

            return (
              <div key={groupName}>
                {groupBy !== 'none' && (
                  <div className="flex items-center justify-between mb-2 gap-3">
                    <h3 className="text-sm font-semibold text-text/80 flex items-center gap-2">
                      {groupBy === 'category' ? (
                        <TagIcon className="w-4 h-4" />
                      ) : (
                        <CalendarIcon className="w-4 h-4" />
                      )}
                      {groupBy === 'date' ? format(new Date(groupName), 'MMMM dd, yyyy') : groupName}
                      <span className="text-xs text-text/50 font-normal">
                        ({groupExpenses.length} {groupExpenses.length === 1 ? 'item' : 'items'})
                      </span>
                    </h3>
                    {groupBy === 'date' && (
                      <span className="px-2 py-1 text-xs font-semibold rounded-full bg-primary/15 text-primary whitespace-nowrap">
                        {formatCurrencyAmount(groupTotal, userHomeCurrency)}
                      </span>
                    )}
                  </div>
                )}

                <div className="space-y-1.5">
                  {groupExpenses.map((expense) => (
                  <div
                    key={expense.id}
                    className="group flex items-center justify-between p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 hover:border-primary/30 transition-all duration-200 cursor-default"
                  >
                    {editingId === expense.id ? (
                       <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3 w-full">
                         <div className="flex flex-col gap-2 flex-1 w-full">
                            <input 
                              type="text"
                              value={editForm.description}
                              onChange={(e) => setEditForm((prev) => ({ ...prev, description: e.target.value }))}
                              className="w-full bg-black/20 border border-white/10 rounded px-2 py-1.5 text-sm text-text focus:border-primary/50 outline-none"
                              placeholder="Description"
                              autoFocus
                            />
                            <div className="flex flex-col gap-2 w-full">
                              <div className="flex flex-col sm:flex-row gap-2 w-full">
                                <div className="flex-1 min-w-0">
                                  <input 
                                    type="number"
                                    value={editForm.originalAmount}
                                    onChange={(e) => {
                                      const value = e.target.value;
                                      setEditForm((prev) => {
                                        const parsed = parseFloat(value);
                                        const numeric = Number.isNaN(parsed) ? 0 : parsed;
                                        return {
                                          ...prev,
                                          originalAmount: numeric,
                                          homePreview: numeric * prev.conversionRate,
                                        };
                                      });
                                    }}
                                    className="w-full bg-black/20 border border-white/10 rounded px-2 py-1.5 text-sm text-text focus:border-primary/50 outline-none"
                                    placeholder="Amount"
                                    step="0.01"
                                    inputMode="decimal"
                                  />
                                  <p className="text-[10px] text-text/60 mt-1">
                                    <CurrencyAmount
                                      amount={editForm.homePreview}
                                      currency={editForm.homeCurrency || 'TWD'}
                                      originalAmount={editForm.originalAmount}
                                      originalCurrency={editForm.currency}
                                      className="text-[10px] text-text/60 mt-1"
                                    />
                                  </p>

                                </div>
                                <select
                                  value={editForm.currency}
                                  onChange={(e) => {
                                    const newCurrency = e.target.value;
                                    setEditForm((prev) => {
                                      const numeric = prev.originalAmount || 0;
                                    let rate = prev.conversionRate || 1;
                                    if (newCurrency === prev.homeCurrency) {
                                      rate = 1;
                                    } else if (expense.exchange_rate && expense.exchange_rate > 0) {
                                      rate = expense.exchange_rate;
                                    }
                                    return {
                                      ...prev,
                                      currency: newCurrency,
                                        conversionRate: rate,
                                        homePreview: numeric * rate,
                                      };
                                    });
                                  }}
                                  className="flex-1 bg-black/20 border border-white/10 rounded px-2 py-1.5 text-sm text-text/80 focus:border-primary/50 outline-none min-w-0"
                                >
                                  {currencyOptions.map((currency) => (
                                    <option key={currency} value={currency}>
                                      {currency}
                                    </option>
                                  ))}
                                  {editForm.currency && !currencyOptions.includes(editForm.currency) && (
                                    <option value={editForm.currency}>{editForm.currency}</option>
                                  )}
                                </select>
                              </div>
                              <select
                                value={editForm.account}
                                onChange={(e) => setEditForm((prev) => ({ ...prev, account: e.target.value }))}
                                className="flex-1 bg-black/20 border border-white/10 rounded px-2 py-1.5 text-sm text-text/80 focus:border-primary/50 outline-none min-w-0"
                              >
                                <option value="">Select account</option>
                                {accountOptions.map((account) => (
                                  <option key={account} value={account}>
                                    {account}
                                  </option>
                                ))}
                                {editForm.account && !accountOptions.includes(editForm.account) && (
                                  <option value={editForm.account}>{editForm.account}</option>
                                )}
                              </select>
                            </div>
                         </div>
                         <div className="flex sm:flex-col items-center gap-1 w-full sm:w-auto pt-2 sm:pt-0">
                            <button 
                              onClick={(e) => saveEditing(expense, e)}
                             disabled={isSaving}
                             className="flex-1 sm:flex-none p-2 rounded-md bg-green-500/10 text-green-400 hover:bg-green-500/20 transition-colors flex items-center justify-center"
                           >
                             <CheckIcon className="w-5 h-5 sm:w-4 sm:h-4" />
                             <span className="sm:hidden ml-2 text-xs font-medium">Save</span>
                           </button>
                           <button 
                             onClick={(e) => cancelEditing(e)}
                             disabled={isSaving}
                             className="flex-1 sm:flex-none p-2 rounded-md bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors flex items-center justify-center"
                           >
                             <XMarkIcon className="w-5 h-5 sm:w-4 sm:h-4" />
                             <span className="sm:hidden ml-2 text-xs font-medium">Cancel</span>
                           </button>
                         </div>
                      </div>
                    ) : (
                      <>
                        <div className="flex items-start gap-3 sm:gap-4 flex-1 min-w-0">
                          {/* Icon */}
                          <div className="flex-shrink-0 w-8 h-8 sm:w-10 sm:h-10 rounded-full flex items-center justify-center bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white transition-colors mt-0.5">
                            <CurrencyDollarIcon className="w-4 h-4 sm:w-5 h-5" />
                          </div>

                          {/* Details Container */}
                          <div className="flex-1 min-w-0">
                            {/* Top Row: Description & Amount (Mobile) */}
                            <div className="flex justify-between items-start gap-2">
                              <p className="text-sm font-medium text-text group-hover:text-primary transition-colors truncate">
                                {expense.description}
                              </p>
                              <div className="sm:hidden text-text group-hover:text-primary transition-colors shrink-0">
                                <CurrencyAmount
                                  amount={expense.home_amount ?? expense.amount}
                                  currency={expense.home_currency || 'TWD'}
                                  originalAmount={expense.original_amount}
                                  originalCurrency={expense.original_currency || expense.currency}
                                  className="items-end"
                                />
                              </div>
                            </div>
                            
                            {/* Metadata Row */}
                            <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1 sm:mt-0.5 text-[10px] sm:text-xs text-text/50">
                              <span className="flex items-center gap-1 shrink-0">
                                <TagIcon className="w-2.5 h-2.5 sm:w-3 h-3" />
                                {expense.category_name || 'Uncategorized'}
                              </span>
                              <span className="flex items-center gap-1 shrink-0">
                                <CalendarIcon className="w-2.5 h-2.5 sm:w-3 h-3" />
                                {format(new Date(expense.expense_date), 'MMM dd, yyyy')}
                              </span>
                              {expense.account && (
                                <span className="flex items-center gap-1 bg-white/10 px-1.5 py-0.5 rounded text-[9px] sm:text-[10px] uppercase tracking-wider font-bold text-primary/80 shrink-0">
                                  {expense.account.toLowerCase().includes('card') ? (
                                    <CreditCardIcon className="w-2.5 h-2.5" />
                                  ) : (
                                    <BanknotesIcon className="w-2.5 h-2.5" />
                                  )}
                                  {expense.account}
                                </span>
                              )}
                              {expense.currency && expense.home_currency && expense.currency !== expense.home_currency && (
                                <span className="flex items-center gap-1 bg-primary/10 px-1.5 py-0.5 rounded text-[9px] sm:text-[10px] uppercase tracking-wider font-bold text-primary/80 shrink-0">
                                  <GlobeAltIcon className="w-2.5 h-2.5" />
                                  {expense.currency}
                                </span>
                              )}
                            </div>
                          </div>
                        </div>

                        {/* Amount & Actions (Desktop) */}
                        <div className="hidden sm:flex items-center gap-4 ml-4">
                          <div className="flex-shrink-0 text-right">
                             <CurrencyAmount 
                                 amount={expense.home_amount ?? expense.amount}
                                 currency={expense.home_currency || 'TWD'}
                                 originalAmount={expense.original_amount}
                                 originalCurrency={expense.original_currency || expense.currency}
                                 className="text-text group-hover:text-primary transition-colors"
                              />
                          </div>
                          
                          {onUpdateExpense && (
                            <button
                              onClick={(e) => startEditing(expense, e)}
                              className="p-2 rounded-lg opacity-0 group-hover:opacity-100 hover:bg-white/10 text-text/40 hover:text-primary transition-all"
                              title="Edit expense"
                            >
                              <PencilSquareIcon className="w-4 h-4" />
                            </button>
                          )}
                          {onDeleteExpense && (
                            <button
                              onClick={(e) => handleDelete(expense, e)}
                              disabled={deletingId === expense.id}
                              className="p-2 rounded-lg opacity-0 group-hover:opacity-100 hover:bg-white/10 text-text/40 hover:text-red-400 transition-all disabled:opacity-50"
                              title="Delete expense"
                            >
                              <TrashIcon className="w-4 h-4" />
                            </button>
                          )}
                        </div>

                        {/* Mobile Action Triggers */}
                        <div className="sm:hidden flex items-center -mr-2">
                          {onUpdateExpense && (
                            <button
                              onClick={(e) => startEditing(expense, e)}
                              className="p-2 text-text/30 hover:text-primary transition-colors"
                            >
                              <PencilSquareIcon className="w-4 h-4" />
                            </button>
                          )}
                          {onDeleteExpense && (
                            <button
                              onClick={(e) => handleDelete(expense, e)}
                              disabled={deletingId === expense.id}
                              className="p-2 text-text/30 hover:text-red-400 transition-colors disabled:opacity-50"
                            >
                              <TrashIcon className="w-4 h-4" />
                            </button>
                          )}
                        </div>
                      </>
                    )}
                  </div>
                  ))}
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Footer Summary */}
      <div className="sticky bottom-0 mt-4 pt-3 pb-1 border-t border-white/10 flex items-center justify-between text-sm bg-background">
        <span className="text-text/60">
          Showing <span className="font-semibold text-text">{processedExpenses.length}</span> of{' '}
          <span className="font-semibold text-text">{expenses.length}</span> expenses
        </span>
        <span className="font-mono font-bold text-text">
          Total: {new Intl.NumberFormat('en-US', { style: 'currency', currency: expenses[0]?.home_currency || 'TWD' }).format(processedExpenses.reduce((sum, e) => sum + (e.home_amount || e.amount), 0))}
        </span>
      </div>
    </div>
  );
}
