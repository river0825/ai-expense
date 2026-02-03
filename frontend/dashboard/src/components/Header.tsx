'use client'

import Link from 'next/link'

interface HeaderProps {
  apiKey: string
  onLogout: () => void
}

export default function Header({ apiKey, onLogout }: HeaderProps) {
  return (
    <header className="bg-slate-800 border-b border-slate-700 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <div>
            <h1 className="text-2xl font-bold text-white">📊 AIExpense Metrics</h1>
            <p className="text-sm text-slate-400">Real-time expense tracking analytics</p>
          </div>
          
          <nav className="hidden md:flex items-center gap-4 border-l border-slate-700 pl-6 h-10">
            <Link 
              href="/chat"
              className="text-slate-400 hover:text-white transition-colors font-medium flex items-center gap-2"
            >
              <span>💬</span> Chat Simulator
            </Link>
          </nav>
        </div>

        <div className="flex items-center gap-4">
          <div className="text-right">
            <p className="text-sm text-slate-400">API Key</p>
            <p className="text-sm font-mono text-slate-300">
              {apiKey.substring(0, 4)}{'*'.repeat(Math.max(0, apiKey.length - 8))}{apiKey.substring(apiKey.length - 4)}
            </p>
          </div>
          <button
            onClick={onLogout}
            className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors text-sm font-medium"
          >
            Logout
          </button>
        </div>
      </div>
    </header>
  )
}
