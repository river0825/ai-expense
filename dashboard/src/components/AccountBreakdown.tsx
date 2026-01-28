import React, { useMemo } from 'react';
import { DashboardCard } from './DashboardCard';
import { CreditCardIcon, BanknotesIcon, WalletIcon } from '@heroicons/react/24/outline';

interface AccountBreakdownProps {
  expenses: { account?: string; amount: number }[];
  className?: string;
}

export function AccountBreakdown({ expenses, className = '' }: AccountBreakdownProps) {
  const accountStats = useMemo(() => {
    const stats: Record<string, number> = {};
    let totalExpenses = 0;

    expenses.forEach(expense => {
      const account = expense.account || 'Cash';
      // Only count expenses (positive amounts in transaction list might be income, but usually expenses are negative or positive depending on model. 
      // In Expense model usually positive amount is expense. In Transaction model mock data, negative is expense, positive is income.
      // Let's assume passed expenses are Expense[] where amount is positive for expense.
      // But wait, if we pass Transaction[], we need to handle signs.
      // The prop says `expenses` so let's check Expense model vs Transaction model usage.
      // In Dashboard page, we might pass transactions or expenses.
      // Let's assume we pass Expense objects which have positive amount.
      // If we pass Transactions, we should pre-process.
      // Let's handle generic { account?: string, amount: number } and assume amount is positive for expense.
      if (expense.amount > 0) {
          stats[account] = (stats[account] || 0) + expense.amount;
          totalExpenses += expense.amount;
      }
    });

    return Object.entries(stats)
      .map(([name, total]) => ({
        name,
        total,
        percentage: totalExpenses > 0 ? (total / totalExpenses) * 100 : 0
      }))
      .sort((a, b) => b.total - a.total);
  }, [expenses]);

  return (
    <DashboardCard title="Account Breakdown" className={className}>
      <div className="space-y-4">
        {accountStats.length === 0 ? (
          <div className="text-center py-8 text-text/40 text-sm">
            No account data available
          </div>
        ) : (
          accountStats.map((account) => (
            <div key={account.name} className="space-y-2 group">
              <div className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2">
                  <div className={`p-1.5 rounded-md ${
                    account.name.toLowerCase().includes('card') 
                      ? 'bg-purple-500/10 text-purple-400' 
                      : 'bg-emerald-500/10 text-emerald-400'
                  }`}>
                    {account.name.toLowerCase().includes('card') ? (
                      <CreditCardIcon className="w-4 h-4" />
                    ) : (
                      <BanknotesIcon className="w-4 h-4" />
                    )}
                  </div>
                  <span className="font-medium text-text group-hover:text-primary transition-colors">
                    {account.name}
                  </span>
                </div>
                <div className="text-right">
                  <p className="font-mono font-bold text-text">${account.total.toFixed(2)}</p>
                </div>
              </div>
              
              {/* Progress Bar */}
              <div className="h-1.5 w-full bg-white/5 rounded-full overflow-hidden">
                <div 
                  className={`h-full rounded-full transition-all duration-500 ${
                    account.name.toLowerCase().includes('card') ? 'bg-purple-500' : 'bg-emerald-500'
                  }`}
                  style={{ width: `${account.percentage}%` }}
                />
              </div>
              
              <div className="flex justify-end">
                <span className="text-xs text-text/40 font-mono">{account.percentage.toFixed(1)}%</span>
              </div>
            </div>
          ))
        )}
      </div>
    </DashboardCard>
  );
}
