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

interface ChartSectionProps {
  metrics: {
    dau: any
    expenses: any
    growth: any
  } | null
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

export default function ChartSection({ metrics }: ChartSectionProps) {
  if (!metrics) return null

  // Prepare DAU data
  const dauData = Array.isArray(metrics.dau) ? metrics.dau.slice(0, 30).map((item: any) => ({
    date: new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    users: item.active_users || 0,
  })) : []

  // Prepare expenses data
  const expensesData = Array.isArray(metrics.expenses) ? metrics.expenses.slice(0, 30).map((item: any) => ({
    date: new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    total: (item.total_expense || 0).toFixed(2),
    count: item.expense_count || 0,
  })) : []

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
      {/* Daily Active Users Chart */}
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

      {/* Expenses Chart */}
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
  )
}
