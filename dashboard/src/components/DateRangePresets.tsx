'use client';

import React from 'react';
import { DateRange } from 'react-day-picker';
import { startOfToday, startOfWeek, startOfMonth, subDays, endOfToday, endOfWeek, endOfMonth } from 'date-fns';
import { DatePreset } from '@/domain/models/Expense';

interface DateRangePresetsProps {
  onSelectPreset: (range: DateRange, preset: DatePreset) => void;
  currentPreset?: DatePreset;
}

export function DateRangePresets({ onSelectPreset, currentPreset }: DateRangePresetsProps) {
  const presets: Array<{ id: DatePreset; label: string; range: () => DateRange }> = [
    {
      id: 'today',
      label: 'Today',
      range: () => ({ from: startOfToday(), to: endOfToday() }),
    },
    {
      id: 'week',
      label: 'This Week',
      range: () => ({ from: startOfWeek(new Date(), { weekStartsOn: 1 }), to: endOfWeek(new Date(), { weekStartsOn: 1 }) }),
    },
    {
      id: 'month',
      label: 'This Month',
      range: () => ({ from: startOfMonth(new Date()), to: endOfMonth(new Date()) }),
    },
    {
      id: 'last7',
      label: 'Last 7 Days',
      range: () => ({ from: subDays(new Date(), 6), to: new Date() }),
    },
    {
      id: 'last30',
      label: 'Last 30 Days',
      range: () => ({ from: subDays(new Date(), 29), to: new Date() }),
    },
    {
      id: 'last90',
      label: 'Last 90 Days',
      range: () => ({ from: subDays(new Date(), 89), to: new Date() }),
    },
  ];

  return (
    <div className="flex flex-wrap gap-2">
      {presets.map((preset) => (
        <button
          key={preset.id}
          onClick={() => onSelectPreset(preset.range(), preset.id)}
          className={`px-3 py-1.5 text-sm font-medium rounded-lg transition-all duration-200 cursor-pointer
            ${
              currentPreset === preset.id
                ? 'bg-primary text-white shadow-md'
                : 'bg-white/5 text-text/70 hover:bg-white/10 hover:text-text border border-white/10'
            }
          `}
        >
          {preset.label}
        </button>
      ))}
    </div>
  );
}
