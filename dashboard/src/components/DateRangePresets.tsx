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

  const handleSelectChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selectedId = e.target.value as DatePreset;
    const preset = presets.find(p => p.id === selectedId);
    if (preset) {
      onSelectPreset(preset.range(), selectedId);
    }
  };

  return (
    <div className="relative">
      <select
        value={currentPreset || 'custom'}
        onChange={handleSelectChange}
        className="w-full sm:w-[180px] bg-white/5 border border-white/10 rounded-lg py-2 pl-3 pr-8 text-sm text-text focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 appearance-none cursor-pointer"
      >
        <option value="custom" disabled>Custom Range</option>
        {presets.map((preset) => (
          <option key={preset.id} value={preset.id} className="bg-surface text-text">
            {preset.label}
          </option>
        ))}
      </select>
      <div className="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none text-text/60">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </div>
    </div>
  );
}
