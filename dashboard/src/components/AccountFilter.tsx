import React from 'react';
import { CreditCardIcon, BanknotesIcon, WalletIcon } from '@heroicons/react/24/outline';

interface AccountFilterProps {
  accounts: string[];
  selectedAccount: string | null;
  onSelectAccount: (account: string | null) => void;
  className?: string;
}

export function AccountFilter({ accounts, selectedAccount, onSelectAccount, className = '' }: AccountFilterProps) {
  return (
    <div className={`flex items-center gap-2 ${className}`}>
      <div className="relative">
        <div className="absolute left-3 top-1/2 -translate-y-1/2 text-text/40 pointer-events-none">
          <WalletIcon className="w-4 h-4" />
        </div>
        <select
          value={selectedAccount || ''}
          onChange={(e) => onSelectAccount(e.target.value || null)}
          className="appearance-none bg-white/5 border border-white/10 rounded-lg pl-9 pr-8 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-primary/50 cursor-pointer hover:bg-white/10 transition-colors"
        >
          <option value="">All Accounts</option>
          {accounts.map((account) => (
            <option key={account} value={account}>
              {account}
            </option>
          ))}
        </select>
        <div className="absolute right-3 top-1/2 -translate-y-1/2 text-text/40 pointer-events-none">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4">
            <path fillRule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clipRule="evenodd" />
          </svg>
        </div>
      </div>
    </div>
  );
}
