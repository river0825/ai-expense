'use client';

import React, { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { DashboardCard } from '@/components/DashboardCard';
import { Sidebar } from '@/components/Sidebar';
import { TopBar } from '@/components/TopBar';
import { 
  ArrowTrendingUpIcon, 
  ArrowTrendingDownIcon,
  CurrencyDollarIcon,
  WalletIcon,
  ShoppingBagIcon
} from '@heroicons/react/24/outline';
import RepositoryFactory from '@/infrastructure/RepositoryFactory';
import { DashboardStats } from '@/domain/models/Stats';
import { Transaction } from '@/domain/models/Transaction';

const STATS_UI = {
  totalBalance: { icon: WalletIcon, color: 'from-blue-500 to-cyan-400' },
  totalIncome: { icon: CurrencyDollarIcon, color: 'from-emerald-500 to-lime-400' },
  totalExpenses: { icon: ShoppingBagIcon, color: 'from-rose-500 to-orange-400' }
};

export default function Dashboard() {
  const t = useTranslations('Dashboard');
  const tStats = useTranslations('Stats');
  
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const statsRepo = RepositoryFactory.getStatsRepository();
        const txRepo = RepositoryFactory.getTransactionRepository();
        
        const [statsData, txData] = await Promise.all([
          statsRepo.getDashboardStats(),
          txRepo.getRecentTransactions(5)
        ]);
        
        setStats(statsData);
        setTransactions(txData);
      } catch (error) {
        console.error('Failed to fetch dashboard data', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading || !stats) {
    return (
      <div className="min-h-screen bg-background font-sans text-text flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  const statKeys = ['totalBalance', 'totalIncome', 'totalExpenses'] as const;

  return (
    <div className="min-h-screen bg-background font-sans text-text selection:bg-primary/30">
      <Sidebar />
      <TopBar />

      <main className="pl-64 pt-20">
        <div className="p-8 max-w-7xl mx-auto space-y-8">
          
          {/* Header Section */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h1 className="text-3xl font-mono font-bold text-text tracking-tight mb-1">
                {t('title')}<span className="text-primary">.</span>{t('subtitle')}
              </h1>
              <p className="text-text/60">{t('welcome')}</p>
            </div>
            <div className="flex gap-3">
               <button className="btn-secondary text-sm py-2 px-4 hover:bg-secondary/10">{t('downloadReport')}</button>
               <button className="btn-primary text-sm py-2 px-4 shadow-lg shadow-cta/25 ring-2 ring-cta/20">{t('addExpense')}</button>
            </div>
          </div>

          {/* Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {statKeys.map((key) => {
              const stat = stats[key];
              const ui = STATS_UI[key];
              const Icon = ui.icon;
              
              return (
                <DashboardCard key={key} className="relative overflow-hidden group">
                   <div className={`absolute -right-6 -top-6 w-24 h-24 rounded-full bg-gradient-to-br ${ui.color} opacity-20 blur-2xl group-hover:opacity-30 transition-opacity duration-500`}></div>
                   <div className="relative z-10">
                     <div className="flex justify-between items-start mb-4">
                       <div className={`p-3 rounded-xl bg-white/5 border border-white/5 text-white ring-1 ring-inset ring-white/10`}>
                         <Icon className="w-6 h-6" />
                       </div>
                       <span className={`flex items-center gap-1 text-sm font-medium px-2 py-1 rounded-full border border-white/5 ${stat.isPositive ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400'}`}>
                         {stat.isPositive ? <ArrowTrendingUpIcon className="w-3 h-3" /> : <ArrowTrendingDownIcon className="w-3 h-3" />}
                         {stat.change}
                       </span>
                     </div>
                     <div className="space-y-1">
                        <p className="text-text/60 text-sm font-medium uppercase tracking-wider">{t(key)}</p>
                        <h3 className="text-3xl font-mono font-bold text-text tracking-tight">{stat.value}</h3>
                     </div>
                   </div>
                </DashboardCard>
              );
            })}
          </div>

          {/* Main Grid: Chart & Transactions */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            
            {/* Chart Section (Mock) */}
            <DashboardCard title={t('revenueAnalytics')} className="lg:col-span-2 min-h-[400px]" 
              action={
                <select className="bg-white/5 border border-white/10 rounded-lg text-sm text-text/80 px-3 py-1 focus:outline-none focus:ring-1 focus:ring-primary/50">
                  <option>{tStats('monthly')}</option>
                  <option>{tStats('weekly')}</option>
                  <option>{tStats('daily')}</option>
                </select>
              }
            >
              <div className="flex flex-col justify-end h-full">
                {/* Mock Chart Visual */}
                <div className="flex items-end justify-between h-64 gap-2 px-2 pb-4">
                  {[40, 65, 45, 80, 55, 90, 70, 85, 60, 75, 50, 95].map((h, i) => (
                    <div key={i} className="w-full bg-white/5 rounded-t-sm relative group">
                      <div 
                        style={{ height: `${h}%` }} 
                        className="w-full absolute bottom-0 bg-gradient-to-t from-primary/20 to-cta/60 rounded-t-sm hover:from-primary/40 hover:to-cta/80 transition-all duration-300"
                      >
                         <div className="opacity-0 group-hover:opacity-100 absolute -top-10 left-1/2 -translate-x-1/2 bg-surface border border-white/10 text-xs px-2 py-1 rounded shadow-xl text-white transition-opacity duration-200">
                           {h}%
                         </div>
                      </div>
                    </div>
                  ))}
                </div>
                {/* X Axis */}
                <div className="flex justify-between px-2 pt-4 border-t border-white/5 text-xs text-text/40 font-mono">
                  <span>JAN</span><span>FEB</span><span>MAR</span><span>APR</span><span>MAY</span><span>JUN</span>
                  <span>JUL</span><span>AUG</span><span>SEP</span><span>OCT</span><span>NOV</span><span>DEC</span>
                </div>
              </div>
            </DashboardCard>

            {/* Recent Transactions */}
            <DashboardCard title={t('recentTransactions')} className="h-full" action={<button className="text-xs text-primary hover:text-primary/80 transition-colors">{t('viewAll')}</button>}>
               <div className="space-y-4">
                 {transactions.map((tx) => (
                   <div key={tx.id} className="group flex items-center justify-between p-3 rounded-xl hover:bg-white/5 border border-transparent hover:border-white/5 transition-all duration-200 cursor-pointer">
                     <div className="flex items-center gap-4">
                       <div className={`w-10 h-10 rounded-full flex items-center justify-center border border-white/5 ${tx.amount.startsWith('+') ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400' }`}>
                          {tx.amount.startsWith('+') ? <ArrowTrendingUpIcon className="w-5 h-5" /> : <ShoppingBagIcon className="w-5 h-5" />}
                       </div>
                       <div>
                         <p className="text-sm font-medium text-text group-hover:text-primary transition-colors">{tx.name}</p>
                         <p className="text-xs text-text/40">{tx.category} • {tx.date}</p>
                       </div>
                     </div>
                     <div className="text-right">
                       <p className={`text-sm font-mono font-semibold ${tx.amount.startsWith('+') ? 'text-emerald-400' : 'text-text'}`}>
                         {tx.amount}
                       </p>
                       <p className={`text-[10px] uppercase font-medium tracking-wide ${tx.status === 'Completed' ? 'text-emerald-500/60' : 'text-amber-500/60'}`}>
                         {tx.status}
                       </p>
                     </div>
                   </div>
                 ))}
               </div>
            </DashboardCard>
          </div>
          
          {/* Quick Actions / Footer */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pb-8">
             <div className="glass-panel p-6 rounded-2xl flex items-center justify-between group cursor-pointer hover:border-cta/30 transition-all">
                <div>
                   <h3 className="text-lg font-bold text-text mb-1 group-hover:text-cta transition-colors">{t('connectBank')}</h3>
                   <p className="text-sm text-text/60">{t('connectBankDesc')}</p>
                </div>
                <div className="w-10 h-10 rounded-full bg-cta/10 flex items-center justify-center text-cta group-hover:scale-110 transition-transform">
                   <ArrowTrendingUpIcon className="w-5 h-5" />
                </div>
             </div>
             <div className="glass-panel p-6 rounded-2xl flex items-center justify-between group cursor-pointer hover:border-primary/30 transition-all">
                <div>
                   <h3 className="text-lg font-bold text-text mb-1 group-hover:text-primary transition-colors">{t('generateReport')}</h3>
                   <p className="text-sm text-text/60">{t('generateReportDesc')}</p>
                </div>
                <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 transition-transform">
                   <DocumentTextIcon className="w-5 h-5" />
                </div>
             </div>
          </div>
          
        </div>
      </main>
    </div>
  );
}

function DocumentTextIcon({ className }: { className?: string }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
    </svg>
  );
}
