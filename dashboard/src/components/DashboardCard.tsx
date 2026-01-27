import React from 'react';

interface DashboardCardProps {
  title?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  action?: React.ReactNode;
}

export function DashboardCard({ title, children, className = '', action }: DashboardCardProps) {
  return (
    <div className={`glass-card rounded-2xl p-6 flex flex-col ${className}`}>
      {(title || action) && (
        <div className="flex items-center justify-between mb-4">
          {title && <div className="text-lg font-mono font-semibold text-text tracking-tight">{title}</div>}
          {action && <div>{action}</div>}
        </div>
      )}
      <div className="flex-1">
        {children}
      </div>
    </div>
  );
}
