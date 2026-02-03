'use client'

import ChatInterface from '@/components/ChatInterface'
import Link from 'next/link'

export default function ChatPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      <header className="bg-slate-800 border-b border-slate-700 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link 
              href="/"
              className="text-slate-400 hover:text-white transition-colors flex items-center gap-2"
            >
              ← Back to Dashboard
            </Link>
            <div className="h-6 w-px bg-slate-700 mx-2"></div>
            <h1 className="text-xl font-bold text-white">💬 Chat Simulator</h1>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="mb-6">
          <h2 className="text-2xl font-bold text-white mb-2">Test Terminal Chat</h2>
          <p className="text-slate-400">
            Interact with the expense bot directly. Expenses added here will be reflected in the dashboard.
          </p>
        </div>
        
        <ChatInterface />
      </main>
    </div>
  )
}
