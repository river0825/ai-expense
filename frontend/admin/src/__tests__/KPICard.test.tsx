import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { KPICard } from '../components/KPICard';
import { Metric } from '../lib/types';

describe('KPICard', () => {
  const mockMetric: Metric = {
    current: 50000,
    previous: 45000,
    delta_percent: 11.11,
  };

  it('renders title and value correctly', () => {
    render(<KPICard title="MRR" metric={mockMetric} prefix="$" />);
    
    expect(screen.getByText('MRR')).toBeDefined();
    expect(screen.getByText('$50,000')).toBeDefined();
  });

  it('renders positive delta correctly', () => {
    render(<KPICard title="MRR" metric={mockMetric} />);
    
    const deltaElement = screen.getByText('11.11%');
    expect(deltaElement).toBeDefined();
    // Check for green color class (implementation detail, but useful for visual regression proxy)
    expect(deltaElement.className).toContain('text-green-500');
  });

  it('renders negative delta correctly', () => {
    const negativeMetric = { ...mockMetric, delta_percent: -5.5 };
    render(<KPICard title="MRR" metric={negativeMetric} />);
    
    const deltaElement = screen.getByText('5.50%');
    expect(deltaElement).toBeDefined();
    expect(deltaElement.className).toContain('text-red-500');
  });

  it('renders inverse delta correctly (churn)', () => {
    const churnMetric = { ...mockMetric, delta_percent: -5.5 };
    render(<KPICard title="Churn" metric={churnMetric} inverse />);
    
    const deltaElement = screen.getByText('5.50%');
    expect(deltaElement).toBeDefined();
    // Negative churn is good -> green
    expect(deltaElement.className).toContain('text-green-500');
  });
});
