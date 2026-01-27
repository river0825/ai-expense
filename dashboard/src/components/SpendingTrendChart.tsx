'use client';

import React from 'react';
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { format } from 'date-fns';
import { TrendDataPoint } from '@/domain/models/Expense';

interface SpendingTrendChartProps {
  data: TrendDataPoint[];
  groupBy: 'day' | 'week' | 'month';
  className?: string;
}

export function SpendingTrendChart({ data, groupBy, className = '' }: SpendingTrendChartProps) {
  // Format data for recharts
  const chartData = data.map(point => ({
    date: groupBy === 'day' 
      ? format(new Date(point.date), 'MMM dd')
      : groupBy === 'week'
      ? `Week of ${format(new Date(point.date), 'MMM dd')}`
      : format(new Date(point.date), 'MMM yyyy'),
    amount: point.amount,
    count: point.count,
  }));

  // Custom tooltip
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-surface/95 backdrop-blur-sm border border-white/10 rounded-lg p-3 shadow-xl">
          <p className="text-sm font-medium text-text mb-1">{payload[0].payload.date}</p>
          <p className="text-lg font-mono font-bold text-primary">
            ${payload[0].value.toFixed(2)}
          </p>
          <p className="text-xs text-text/60 mt-1">
            {payload[0].payload.count} transaction{payload[0].payload.count !== 1 ? 's' : ''}
          </p>
        </div>
      );
    }
    return null;
  };

  if (data.length === 0) {
    return (
      <div className={`flex items-center justify-center h-full ${className}`}>
        <p className="text-text/40 text-sm">No trend data available for this period</p>
      </div>
    );
  }

  return (
    <div className={`w-full ${className}`}>
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="amountGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#3B82F6" stopOpacity={0}/>
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#ffffff10" />
          <XAxis 
            dataKey="date" 
            stroke="#94a3b8"
            style={{ fontSize: '12px', fontFamily: 'Fira Sans, sans-serif' }}
            tick={{ fill: '#94a3b8' }}
          />
          <YAxis 
            stroke="#94a3b8"
            style={{ fontSize: '12px', fontFamily: 'Fira Code, monospace' }}
            tick={{ fill: '#94a3b8' }}
            tickFormatter={(value) => `$${value}`}
          />
          <Tooltip content={<CustomTooltip />} />
          <Area
            type="monotone"
            dataKey="amount"
            stroke="#3B82F6"
            strokeWidth={2}
            fill="url(#amountGradient)"
            activeDot={{ r: 6, fill: '#3B82F6', stroke: '#1E293B', strokeWidth: 2 }}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
