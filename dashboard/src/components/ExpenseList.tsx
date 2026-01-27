'use client';

import React, { useState, useMemo } from 'react';
import { format } from 'date-fns';
import { Expense } from '@/domain/models/Expense';
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
  XMarkIcon
} from '@heroicons/react/24/outline';

interface ExpenseListProps {
  expenses: Expense[];
  onCategoryFilter?: (categoryName: string | null) => void;
  onUpdateExpense?: (expense: Expense) => Promise<void>;
  className?: string;
}

type SortField = 'date' | 'amount' | 'category';
type SortDirection = 'asc' | 'desc';

export function ExpenseList({ expenses, onCategoryFilter, onUpdateExpense, className = '' }: ExpenseListProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [sortField, setSortField] = useState<SortField>('date');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [groupBy, setGroupBy] = useState<'none' | 'category' | 'date'>('none');
  
  // Editing state
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<{description: string; amount: string}>({ description: '', amount: '' });
  const [isSaving, setIsSaving] = useState(false);

  const startEditing = (expense: Expense, e: React.MouseEvent) => {
    e.stopPropagation();
    setEditingId(expense.id);
    setEditForm({
      description: expense.description,
      amount: expense.amount.toString()
    });
  };

  const cancelEditing = (e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    setEditingId(null);
    setEditForm({ description: '', amount: '' });
  };

  const saveEditing = async (originalExpense: Expense, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!onUpdateExpense) return;
    
    try {
      setIsSaving(true);
      const updatedExpense: Expense = {
        ...originalExpense,
        description: editForm.description,
        amount: parseFloat(editForm.amount) || 0
      };
      await onUpdateExpense(updatedExpense);
      setEditingId(null);
    } catch (error) {
      console.error('Failed to update expense', error);
    } finally {
      setIsSaving(false);
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
      {/* Controls */}
      <div className="mb-4 space-y-3">
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
          Object.entries(groupedExpenses).map(([groupName, groupExpenses]) => (
            <div key={groupName}>
              {groupBy !== 'none' && (
                <h3 className="text-sm font-semibold text-text/80 mb-2 flex items-center gap-2">
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
              )}

              <div className="space-y-1.5">
                {groupExpenses.map((expense) => (
                  <div
                    key={expense.id}
                    className="group flex items-center justify-between p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 hover:border-primary/30 transition-all duration-200 cursor-default"
                  >
                    {editingId === expense.id ? (
                      <div className="flex items-center gap-4 flex-1 w-full">
                         <div className="flex-1 min-w-0 flex gap-2">
                           <input 
                             type="text"
                             value={editForm.description}
                             onChange={(e) => setEditForm({...editForm, description: e.target.value})}
                             className="flex-1 bg-black/20 border border-white/10 rounded px-2 py-1 text-sm text-text focus:border-primary/50 outline-none"
                             placeholder="Description"
                             autoFocus
                           />
                           <input 
                             type="number"
                             value={editForm.amount}
                             onChange={(e) => setEditForm({...editForm, amount: e.target.value})}
                             className="w-24 bg-black/20 border border-white/10 rounded px-2 py-1 text-sm text-text focus:border-primary/50 outline-none text-right"
                             placeholder="Amount"
                             step="0.01"
                           />
                         </div>
                         <div className="flex items-center gap-1">
                           <button 
                             onClick={(e) => saveEditing(expense, e)}
                             disabled={isSaving}
                             className="p-1.5 rounded-md bg-green-500/10 text-green-400 hover:bg-green-500/20 transition-colors"
                           >
                             <CheckIcon className="w-4 h-4" />
                           </button>
                           <button 
                             onClick={(e) => cancelEditing(e)}
                             disabled={isSaving}
                             className="p-1.5 rounded-md bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                           >
                             <XMarkIcon className="w-4 h-4" />
                           </button>
                         </div>
                      </div>
                    ) : (
                      <>
                        <div className="flex items-center gap-4 flex-1 min-w-0">
                          {/* Icon */}
                          <div className="flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white transition-colors">
                            <CurrencyDollarIcon className="w-5 h-5" />
                          </div>

                          {/* Details */}
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-text group-hover:text-primary transition-colors truncate">
                              {expense.description}
                            </p>
                            <div className="flex items-center gap-3 mt-0.5 text-xs text-text/50">
                              <span className="flex items-center gap-1">
                                <TagIcon className="w-3 h-3" />
                                {expense.category_name || 'Uncategorized'}
                              </span>
                              <span className="flex items-center gap-1">
                                <CalendarIcon className="w-3 h-3" />
                                {format(new Date(expense.expense_date), 'MMM dd, yyyy')}
                              </span>
                            </div>
                          </div>
                        </div>

                        {/* Amount & Actions */}
                        <div className="flex items-center gap-4 ml-4">
                          <div className="flex-shrink-0 text-right">
                            <p className="text-base font-mono font-bold text-text group-hover:text-primary transition-colors">
                              ${expense.amount.toFixed(2)}
                            </p>
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
                        </div>
                      </>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </div>

      {/* Footer Summary */}
      <div className="mt-4 pt-3 border-t border-white/10 flex items-center justify-between text-sm">
        <span className="text-text/60">
          Showing <span className="font-semibold text-text">{processedExpenses.length}</span> of{' '}
          <span className="font-semibold text-text">{expenses.length}</span> expenses
        </span>
        <span className="font-mono font-bold text-text">
          Total: ${processedExpenses.reduce((sum, e) => sum + e.amount, 0).toFixed(2)}
        </span>
      </div>
    </div>
  );
}
