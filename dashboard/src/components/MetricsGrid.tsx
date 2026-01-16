'use client'

interface MetricsGridProps {
  metrics: {
    dau: any
    expenses: any
    growth: any
  } | null
}

interface MetricCard {
  title: string
  value: string | number
  subtitle?: string
  icon: string
  trend?: { value: number; positive: boolean }
}

export default function MetricsGrid({ metrics }: MetricsGridProps) {
  if (!metrics) return null

  const growth = metrics.growth

  const metricCards: MetricCard[] = [
    {
      title: 'Total Users',
      value: growth.total_users || 0,
      icon: '👥',
    },
    {
      title: 'New Users Today',
      value: growth.new_users_today || 0,
      icon: '✨',
      trend: { value: growth.daily_growth_percent || 0, positive: true },
    },
    {
      title: 'Users This Week',
      value: growth.new_users_this_week || 0,
      icon: '📈',
    },
    {
      title: 'Users This Month',
      value: growth.new_users_this_month || 0,
      icon: '📅',
    },
    {
      title: 'Total Expenses',
      value: `$${(growth.total_expenses || 0).toFixed(2)}`,
      icon: '💰',
    },
    {
      title: 'Avg per User',
      value: `$${(growth.average_expense_per_user || 0).toFixed(2)}`,
      icon: '📊',
    },
  ]

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
      {metricCards.map((card, index) => (
        <div
          key={index}
          className="bg-slate-800 rounded-lg border border-slate-700 p-6 hover:border-slate-600 transition-colors"
        >
          <div className="flex items-start justify-between">
            <div>
              <p className="text-slate-400 text-sm font-medium">{card.title}</p>
              <p className="text-2xl font-bold text-white mt-2">{card.value}</p>
              {card.subtitle && (
                <p className="text-xs text-slate-500 mt-1">{card.subtitle}</p>
              )}
            </div>
            <span className="text-3xl">{card.icon}</span>
          </div>

          {card.trend && (
            <div className={`mt-4 text-sm font-medium ${card.trend.positive ? 'text-green-400' : 'text-red-400'}`}>
              {card.trend.positive ? '↑' : '↓'} {card.trend.value.toFixed(2)}%
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
