'use client'

import { useEffect, useState } from 'react'
import axios from 'axios'
import MetricsGrid from '@/components/MetricsGrid'
import ChartSection from '@/components/ChartSection'
import Header from '@/components/Header'

interface MetricsData {
  dau: any
  expenses: any
  growth: any
}

interface AICostMetrics {
  summary: {
    total_calls: number
    total_input_tokens: number
    total_output_tokens: number
    total_tokens: number
    total_cost: number
    currency: string
  }
  daily_stats: any[]
  by_operation: any[]
  top_users: any[]
}

export default function Dashboard() {
  const [metrics, setMetrics] = useState<MetricsData | null>(null)
  const [aiCosts, setAiCosts] = useState<AICostMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [apiKey, setApiKey] = useState('')
  const [showKeyInput, setShowKeyInput] = useState(true)

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

  const fetchMetrics = async (key: string) => {
    try {
      setLoading(true)
      setError(null)

      const headers = { 'X-API-Key': key }

      const [dauRes, expensesRes, growthRes, aiCostsRes] = await Promise.all([
        axios.get(`${apiUrl}/api/metrics/dau`, { headers }),
        axios.get(`${apiUrl}/api/metrics/expenses-summary`, { headers }),
        axios.get(`${apiUrl}/api/metrics/growth`, { headers }),
        axios.get(`${apiUrl}/api/metrics/ai-costs`, { headers }).catch(() => ({ data: { data: null } })),
      ])

      setMetrics({
        dau: dauRes.data.data,
        expenses: expensesRes.data.data,
        growth: growthRes.data.data,
      })

      setAiCosts(aiCostsRes.data.data)

      setShowKeyInput(false)
      localStorage.setItem('apiKey', key)
    } catch (err) {
      setError('Failed to fetch metrics. Check API key and server connection.')
      console.error('Error fetching metrics:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const savedKey = localStorage.getItem('apiKey')
    if (savedKey) {
      setApiKey(savedKey)
      fetchMetrics(savedKey)
    } else {
      setLoading(false)
    }
  }, [])

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (apiKey.trim()) {
      fetchMetrics(apiKey)
    }
  }

  if (showKeyInput) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <div className="bg-slate-800 rounded-lg shadow-xl p-8 border border-slate-700">
            <h1 className="text-3xl font-bold text-white mb-2">AIExpense</h1>
            <p className="text-slate-400 mb-6">Metrics Dashboard</p>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  API Key
                </label>
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="Enter your admin API key"
                  className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
                />
              </div>

              {error && (
                <div className="p-3 bg-red-900 border border-red-700 rounded-lg text-red-200 text-sm">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading || !apiKey.trim()}
                className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-600 text-white font-medium rounded-lg transition-colors"
              >
                {loading ? 'Loading...' : 'View Metrics'}
              </button>

              <div className="pt-4 border-t border-slate-700">
                <p className="text-xs text-slate-400">
                  Default API Key: <code className="bg-slate-700 px-2 py-1 rounded">admin_key</code>
                </p>
              </div>
            </form>
          </div>
        </div>
      </div>
    )
  }

  if (error && !metrics) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-6">
        <div className="max-w-6xl mx-auto">
          <button
            onClick={() => setShowKeyInput(true)}
            className="mb-6 px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors"
          >
            ← Back
          </button>
          <div className="p-6 bg-red-900 border border-red-700 rounded-lg text-red-200">
            {error}
          </div>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center">
        <div className="text-white">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p>Loading metrics...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      <Header apiKey={apiKey} onLogout={() => setShowKeyInput(true)} />

      <main className="max-w-7xl mx-auto px-4 py-8">
        <MetricsGrid metrics={metrics} aiCosts={aiCosts} />
        <ChartSection metrics={metrics} aiCosts={aiCosts} />
      </main>
    </div>
  )
}
