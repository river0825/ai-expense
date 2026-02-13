import React from 'react';
import { cn } from '@/lib/utils';
import { Period } from '@/lib/types';

interface PeriodFilterProps {
  value: Period;
  onChange: (period: Period) => void;
  className?: string;
}

export const PeriodFilter: React.FC<PeriodFilterProps> = ({
  value,
  onChange,
  className,
}) => {
  const periods: { label: string; value: Period }[] = [
    { label: '7 Days', value: '7d' },
    { label: '30 Days', value: '30d' },
    { label: '90 Days', value: '90d' },
  ];

  return (
    <div className={cn("inline-flex bg-muted p-1 rounded-lg", className)}>
      {periods.map((period) => (
        <button
          key={period.value}
          onClick={() => onChange(period.value)}
          data-testid={`period-filter-${period.value}`}
          className={cn(
            "px-3 py-1.5 text-sm font-medium rounded-md transition-all",
            value === period.value
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground hover:bg-background/50"
          )}
        >
          {period.label}
        </button>
      ))}
    </div>
  );
};
