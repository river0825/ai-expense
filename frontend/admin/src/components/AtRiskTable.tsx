import React from 'react';
import { cn } from '@/lib/utils';
import { AtRiskAccount } from '@/lib/types';

interface AtRiskTableProps {
  accounts: AtRiskAccount[];
  className?: string;
}

export const AtRiskTable: React.FC<AtRiskTableProps> = ({ accounts, className }) => {
  return (
    <div className={cn("overflow-x-auto rounded-lg border border-border", className)}>
      <table className="w-full text-sm text-left" data-testid="at-risk-table">
        <thead className="bg-muted text-muted-foreground uppercase text-xs font-semibold">
          <tr>
            <th className="px-6 py-3">Account Name</th>
            <th className="px-6 py-3">MRR</th>
            <th className="px-6 py-3">Health Score</th>
            <th className="px-6 py-3">Risk Factor</th>
            <th className="px-6 py-3">Last Activity</th>
            <th className="px-6 py-3">Action</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border bg-card">
          {accounts.length === 0 ? (
            <tr>
              <td colSpan={6} className="px-6 py-4 text-center text-muted-foreground">
                No at-risk accounts found.
              </td>
            </tr>
          ) : (
            accounts.map((account) => (
              <tr key={account.id} className="hover:bg-muted/50 transition-colors">
                <td className="px-6 py-4 font-medium text-foreground">{account.name}</td>
                <td className="px-6 py-4">${account.mrr.toLocaleString()}</td>
                <td className="px-6 py-4">
                  <span className={cn(
                    "px-2 py-1 rounded-full text-xs font-medium",
                    account.health_score < 40 ? "bg-red-500/10 text-red-500" :
                    account.health_score < 70 ? "bg-yellow-500/10 text-yellow-500" :
                    "bg-green-500/10 text-green-500"
                  )}>
                    {account.health_score}
                  </span>
                </td>
                <td className="px-6 py-4 text-muted-foreground">{account.risk_factor}</td>
                <td className="px-6 py-4 text-muted-foreground">
                  {new Date(account.last_activity).toLocaleDateString()}
                </td>
                <td className="px-6 py-4">
                  <button 
                    className="text-primary hover:underline font-medium"
                    onClick={() => console.log(`View details for ${account.id}`)}
                  >
                    View Details
                  </button>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
};
