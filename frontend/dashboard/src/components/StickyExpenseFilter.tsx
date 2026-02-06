'use client';

import React from 'react';
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { AccountFilter } from './AccountFilter';
import { format } from 'date-fns';

export interface StickyExpenseFilterProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  groupBy: 'date' | 'category';
  onGroupByChange: (groupBy: 'date' | 'category') => void;
  selectedDate: Date;
  onDateSelect: (date: Date) => void;
  selectedAccount: string | null;
  onAccountSelect: (account: string | null) => void;
  accounts: string[];
}

export function StickyExpenseFilter({
  searchQuery,
  onSearchChange,
  groupBy,
  onGroupByChange,
  selectedDate,
  onDateSelect,
  selectedAccount,
  onAccountSelect,
  accounts,
}: StickyExpenseFilterProps) {
  return (
    <div
      data-testid="sticky-filter"
      className="sticky top-0 z-10 bg-background/95 backdrop-blur-sm border-b border-white/10 p-4 space-y-3"
    >
      {/* Search Input */}
      <div className="relative">
        <div className="absolute left-3 top-1/2 -translate-y-1/2 text-text/40 pointer-events-none">
          <MagnifyingGlassIcon className="w-4 h-4" />
        </div>
        <input
          type="text"
          placeholder="Search expenses..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="w-full bg-white/5 border border-white/10 rounded-lg pl-9 pr-4 py-2 text-sm text-text placeholder:text-text/40 focus:outline-none focus:ring-1 focus:ring-primary/50 hover:bg-white/10 transition-colors"
        />
      </div>

      {/* Filter Controls Row */}
      <div className="flex items-center gap-3">
        {/* Group By Buttons */}
        <div className="flex items-center gap-1 px-2 py-1 bg-white/5 rounded-lg border border-white/10">
          <button
            onClick={() => onGroupByChange('date')}
            className={`px-3 py-1 rounded text-sm font-medium transition-colors ${
              groupBy === 'date'
                ? 'bg-primary/20 text-primary'
                : 'text-text/60 hover:text-text'
            }`}
          >
            Date
          </button>
          <button
            onClick={() => onGroupByChange('category')}
            className={`px-3 py-1 rounded text-sm font-medium transition-colors ${
              groupBy === 'category'
                ? 'bg-primary/20 text-primary'
                : 'text-text/60 hover:text-text'
            }`}
          >
            Category
          </button>
        </div>

        {/* Date Picker */}
        <input
          type="date"
          value={format(selectedDate, 'yyyy-MM-dd')}
          onChange={(e) => onDateSelect(new Date(e.target.value))}
          className="bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-primary/50 hover:bg-white/10 transition-colors cursor-pointer"
        />

        {/* Account Filter */}
        <AccountFilter
          accounts={accounts}
          selectedAccount={selectedAccount}
          onSelectAccount={onAccountSelect}
          className="ml-auto"
        />
      </div>
    </div>
  );
}
