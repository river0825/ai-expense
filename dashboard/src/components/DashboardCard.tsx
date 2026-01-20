import React from 'react';

interface DashboardCardProps {
  title?: string;
  children: React.ReactNode;
  className?: string;
  action?: React.ReactNode;
}

export function DashboardCard({ title, children, className = '', action }: DashboardCardProps) {
  return (
    <div className={`glass-card rounded-2xl p-6 flex flex-col ${className}`}>
      {(title || action) && (
        <div className="flex items-center justify-between mb-4">
          {title && <h3 className="text-lg font-mono font-semibold text-text tracking-tight">{title}</h3>}
          {action && <div>{action}</div>}
        </div>
      )}
      <div className="flex-1">
        {children}
      </div>
    </div>
  );
}
