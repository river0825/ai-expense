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

  const isDifferent = originalCurrency && currency !== originalCurrency && originalAmount !== undefined;
  const showOriginal = isDifferent && !hideOriginal;

  return (
    <div className={`flex flex-col items-end ${className}`}>
      <span className="font-mono font-bold whitespace-nowrap">
        {formatCurrency(amount, currency)}
      </span>
      {showOriginal && (
        <span className="text-xs text-text/50 whitespace-nowrap">
          ≈ {new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: originalCurrency,
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          }).format(originalAmount)}
        </span>
      )}
    </div>
  );
}
