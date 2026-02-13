import React from 'react';
import { cn } from '@/lib/utils';

interface TrendChartProps {
  data?: number[]; // Array of values
  color?: string;
  className?: string;
  height?: number;
}

export const TrendChart: React.FC<TrendChartProps> = ({
  data = [],
  color = 'currentColor',
  className,
  height = 40,
}) => {
  // If no data, generate some random-looking trend based on nothing (placeholder)
  // or just return null.
  // Let's generate a fake sparkline if data is empty for visual demo purposes,
  // but in production we'd want real data.
  // For this task, I'll assume data might be passed, or I'll generate a simple sine wave.
  
  const points = data.length > 0 ? data : [10, 15, 13, 17, 14, 20, 25, 22, 30];
  
  const max = Math.max(...points);
  const min = Math.min(...points);
  const range = max - min || 1;
  
  const width = 100; // Viewbox width
  const strokeWidth = 2;
  
  const polylinePoints = points.map((value, index) => {
    const x = (index / (points.length - 1)) * width;
    const y = height - ((value - min) / range) * height;
    return `${x},${y}`;
  }).join(' ');

  return (
    <div className={cn("w-full overflow-hidden", className)} style={{ height }} data-testid="trend-chart">
      <svg
        width="100%"
        height="100%"
        viewBox={`0 0 ${width} ${height}`}
        preserveAspectRatio="none"
        className="overflow-visible"
      >
        <polyline
          points={polylinePoints}
          fill="none"
          stroke={color}
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          strokeLinejoin="round"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </div>
  );
};
