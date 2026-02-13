import React from 'react';
import { cn } from '@/lib/utils';

interface KPIGridProps {
  children: React.ReactNode;
  className?: string;
}

export const KPIGrid: React.FC<KPIGridProps> = ({ children, className }) => {
  return (
    <div className={cn("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4", className)}>
      {children}
    </div>
  );
};
