import React from 'react';
import { ArrowUpIcon, ArrowDownIcon, MinusIcon } from 'lucide-react';
import { cn } from '@/lib/utils'; // Assuming utils exists or I'll create it
import { Metric } from '@/lib/types';

interface KPICardProps {
  title: string;
  metric: Metric;
  prefix?: string;
  suffix?: string;
  className?: string;
  inverse?: boolean; // If true, negative delta is good (e.g. churn)
}

export const KPICard: React.FC<KPICardProps> = ({
  title,
  metric,
  prefix = '',
  suffix = '',
  className,
  inverse = false,
}) => {
  const { current, delta_percent } = metric;
  
  const isPositive = delta_percent > 0;
  const isNegative = delta_percent < 0;
  const isNeutral = delta_percent === 0;

  // Determine color based on delta and inverse flag
  // Normal: Positive = Green, Negative = Red
  // Inverse: Positive = Red, Negative = Green
  let deltaColor = 'text-gray-500';
  if (isPositive) {
    deltaColor = inverse ? 'text-red-500' : 'text-green-500';
  } else if (isNegative) {
    deltaColor = inverse ? 'text-green-500' : 'text-red-500';
  }

  return (
    <div 
      className={cn("bg-card border border-border rounded-lg p-6 shadow-sm", className)}
      data-testid={`kpi-card-${title.toLowerCase().replace(/\s+/g, '-')}`}
    >
      <h3 className="text-sm font-medium text-muted-foreground mb-2">{title}</h3>
      <div className="flex items-baseline justify-between">
        <div className="text-2xl font-bold">
          {prefix}{current.toLocaleString()}{suffix}
        </div>
        <div className={cn("flex items-center text-sm font-medium", deltaColor)}>
          {isPositive && <ArrowUpIcon className="w-4 h-4 mr-1" />}
          {isNegative && <ArrowDownIcon className="w-4 h-4 mr-1" />}
          {isNeutral && <MinusIcon className="w-4 h-4 mr-1" />}
          {Math.abs(delta_percent).toFixed(2)}%
        </div>
      </div>
      <p className="text-xs text-muted-foreground mt-1">vs previous period</p>
    </div>
  );
};
