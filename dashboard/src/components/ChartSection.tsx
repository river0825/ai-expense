'use client'

import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'

interface AICostDailyStats {
  date: string
  calls: number
  input_tokens: number
  output_tokens: number
  total_tokens: number
  cost: number
}

interface AICostByOperation {
  operation: string
  calls: number
  total_tokens: number
  percent: number
}

interface AICostMetrics {
  summary: any
  daily_stats: AICostDailyStats[]
  by_operation: AICostByOperation[]
  top_users: any[]
}

interface ChartSectionProps {
  metrics: {
    dau: any
    expenses: any
    growth: any
  } | null
  aiCosts?: AICostMetrics | null
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

export default function ChartSection({ metrics, aiCosts }: ChartSectionProps) {
  if (!metrics) return null

  const dauData = Array.isArray(metrics.dau) ? metrics.dau.slice(0, 30).map((item: any) => ({
    date: new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    users: item.active_users || 0,
  })) : []

  const expensesData = Array.isArray(metrics.expenses) ? metrics.expenses.slice(0, 30).map((item: any) => ({
    date: new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    total: (item.total_expense || 0).toFixed(2),
    count: item.expense_count || 0,
  })) : []

  const aiDailyData = Array.isArray(aiCosts?.daily_stats) ? aiCosts.daily_stats.map((item) => ({
    date: new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    tokens: item.total_tokens || 0,
    calls: item.calls || 0,
  })) : []

  const aiOperationData = Array.isArray(aiCosts?.by_operation) ? aiCosts.by_operation.map((item) => ({
    name: item.operation.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase()),
    value: item.total_tokens || 0,
    percent: item.percent || 0,
  })) : []

  return (
    <>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <h2 className="text-lg font-bold text-white mb-4">📈 Daily Active Users</h2>
          {dauData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={dauData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="date" stroke="#94a3b8" />
                <YAxis stroke="#94a3b8" />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #475569' }}
                  labelStyle={{ color: '#e2e8f0' }}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="users"
                  stroke="#3b82f6"
                  dot={{ fill: '#3b82f6' }}
                  activeDot={{ r: 5 }}
                  name="Active Users"
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[300px] flex items-center justify-center text-slate-400">
              No data available
            </div>
          )}
        </div>

        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <h2 className="text-lg font-bold text-white mb-4">💰 Daily Expenses</h2>
          {expensesData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={expensesData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="date" stroke="#94a3b8" />
                <YAxis stroke="#94a3b8" />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #475569' }}
                  labelStyle={{ color: '#e2e8f0' }}
                />
                <Legend />
                <Bar dataKey="total" fill="#10b981" name="Total Expenses ($)" />
                <Bar dataKey="count" fill="#3b82f6" name="Transaction Count" />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[300px] flex items-center justify-center text-slate-400">
              No data available
            </div>
          )}
        </div>
      </div>

      {(aiDailyData.length > 0 || aiOperationData.length > 0) && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
            <h2 className="text-lg font-bold text-white mb-4">🤖 Daily AI Token Usage</h2>
            {aiDailyData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={aiDailyData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="date" stroke="#94a3b8" />
                  <YAxis stroke="#94a3b8" />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #475569' }}
                    labelStyle={{ color: '#e2e8f0' }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="tokens"
                    stroke="#8b5cf6"
                    dot={{ fill: '#8b5cf6' }}
                    activeDot={{ r: 5 }}
                    name="Total Tokens"
                  />
                  <Line
                    type="monotone"
                    dataKey="calls"
                    stroke="#f59e0b"
                    dot={{ fill: '#f59e0b' }}
                    activeDot={{ r: 5 }}
                    name="API Calls"
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[300px] flex items-center justify-center text-slate-400">
                No AI usage data available
              </div>
            )}
          </div>

          <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
            <h2 className="text-lg font-bold text-white mb-4">📊 AI Usage by Operation</h2>
            {aiOperationData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={aiOperationData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }) => `${name} (${percent.toFixed(1)}%)`}
                    outerRadius={100}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {aiOperationData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #475569' }}
                    formatter={(value: number) => [value.toLocaleString() + ' tokens', 'Usage']}
                  />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[300px] flex items-center justify-center text-slate-400">
                No operation data available
              </div>
            )}
          </div>
        </div>
      )}
    </>
  )
}
