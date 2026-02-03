'use client'

interface AICostSummary {
  total_calls: number
  total_input_tokens: number
  total_output_tokens: number
  total_tokens: number
  total_cost: number
  currency: string
}

interface AICostMetrics {
  summary: AICostSummary
  daily_stats: any[]
  by_operation: any[]
  top_users: any[]
}

interface MetricsGridProps {
  metrics: {
    dau: any
    expenses: any
    growth: any
  } | null
  aiCosts?: AICostMetrics | null
  currency?: string
}

interface MetricCard {
  title: string
  value: string | number
  subtitle?: string
  icon: string
  trend?: { value: number; positive: boolean }
}

function formatCurrency(num: number, currency: string = 'USD'): string {
  return new Intl.NumberFormat('en-US', { 
    style: 'currency', 
    currency: currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(num)
}

function formatNumber(num: number): string {
  return num.toLocaleString()
}

export default function MetricsGrid({ metrics, aiCosts, currency = 'USD' }: MetricsGridProps) {
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
      value: formatCurrency(growth.total_expenses || 0, currency),
      icon: '💰',
    },
    {
      title: 'Avg per User',
      value: formatCurrency(growth.average_expense_per_user || 0, currency),
      icon: '📊',
    },
  ]

  if (aiCosts?.summary) {
    metricCards.push(
      {
        title: 'Total AI Calls',
        value: formatNumber(aiCosts.summary.total_calls || 0),
        icon: '🤖',
      },
      {
        title: 'Total Tokens',
        value: formatNumber(aiCosts.summary.total_tokens || 0),
        subtitle: `In: ${formatNumber(aiCosts.summary.total_input_tokens || 0)} | Out: ${formatNumber(aiCosts.summary.total_output_tokens || 0)}`,
        icon: '📊',
      },
      {
        title: 'AI Cost',
        value: `$${(aiCosts.summary.total_cost || 0).toFixed(4)}`,
        subtitle: aiCosts.summary.currency || 'USD',
        icon: '💵',
      }
    )
  }

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
