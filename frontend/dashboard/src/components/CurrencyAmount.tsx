import React from 'react';

interface CurrencyAmountProps {
  amount: number;
  currency: string;
  originalAmount?: number;
  originalCurrency?: string;
  className?: string;
  hideOriginal?: boolean;
}

export function CurrencyAmount({ 
  amount, 
  currency, 
  originalAmount, 
  originalCurrency,
  className = '',
  hideOriginal = false
}: CurrencyAmountProps) {
  
  const formatCurrency = (val: number, curr: string) => {
    try {
      if (curr === 'TWD') {
        const number = new Intl.NumberFormat('zh-TW', {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2,
        }).format(val);
        return `NT$${number}`;
      }

      return new Intl.NumberFormat('en-US', { 
        style: 'currency', 
        currency: curr,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(val);
    } catch (e) {
      // Fallback for invalid currency codes
      return `${curr} ${val.toFixed(2)}`;
    }
  };

  const mainAmount = originalAmount ?? amount;
  const mainCurrency = originalCurrency || currency || 'TWD';
  const showHome = !hideOriginal && currency && mainCurrency !== currency;

  return (
    <div className={`flex flex-col items-end ${className}`}>
      <span className="font-mono font-bold whitespace-nowrap">
        {formatCurrency(mainAmount, mainCurrency)}
      </span>
      {showHome && (
        <span className="text-xs text-text/50 whitespace-nowrap">
          ≈ {formatCurrency(amount, currency)}
        </span>
      )}
    </div>
  );
}
