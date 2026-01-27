'use client';

import React, { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { subDays, format } from 'date-fns';
import { DateRange } from 'react-day-picker';
import { DatePickerWithRange } from '@/components/ui/date-range-picker';
import { DateRangePresets } from '@/components/DateRangePresets';
import { ExpenseList } from '@/components/ExpenseList';
import { SpendingTrendChart } from '@/components/SpendingTrendChart';
import { DashboardCard } from '@/components/DashboardCard';
import { Sidebar } from '@/components/Sidebar';
import { TopBar } from '@/components/TopBar';
import RepositoryFactory from '@/infrastructure/RepositoryFactory';
import { Expense, CategoryTotal, TrendDataPoint, DatePreset } from '@/domain/models/Expense';
import { ExpenseReport } from '@/domain/models/Report';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { 
  CurrencyDollarIcon, 
  CalendarIcon, 
  ArrowTrendingUpIcon,
  ListBulletIcon,
  TagIcon,
  ChartBarIcon
} from '@heroicons/react/24/outline';

const COLORS = ['#3B82F6', '#60A5FA', '#F97316', '#FBBF24', '#34D399', '#A78BFA', '#F472B6', '#FB923C'];

export default function ReportPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const urlToken = searchParams.get('token');
  const t = useTranslations('Dashboard');
  
  // State
  const [report, setReport] = useState<ExpenseReport | null>(null);
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [categoryTotals, setCategoryTotals] = useState<CategoryTotal[]>([]);
  const [trendData, setTrendData] = useState<TrendDataPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [date, setDate] = useState<DateRange | undefined>({
    from: subDays(new Date(), 30),
    to: new Date(),
  });
  const [currentPreset, setCurrentPreset] = useState<DatePreset>('last30');
  const [trendGroupBy, setTrendGroupBy] = useState<'day' | 'week' | 'month'>('day');

  // Token management
  useEffect(() => {
    if (urlToken) {
      localStorage.setItem('report_token', urlToken);
    }
  }, [urlToken]);

  const getToken = () => {
    return urlToken || localStorage.getItem('report_token');
  };

  // Fetch data
  useEffect(() => {
    const fetchData = async () => {
      const token = getToken();
      if (!token) {
        setError('Missing access token');
        setLoading(false);
        return;
      }

      setLoading(true);
      try {
        const reportRepo = RepositoryFactory.getReportRepository();
        const expenseRepo = RepositoryFactory.getExpenseRepository();
        
        // Fetch all data in parallel
        const [reportData, expensesData, categoryData, trendDataRaw] = await Promise.all([
          reportRepo.getReportSummary(token, date?.from, date?.to),
          expenseRepo.getExpenses(token, date?.from, date?.to),
          expenseRepo.getCategoryTotals(token, date?.from, date?.to),
          date?.from && date?.to 
            ? expenseRepo.getTrendData(token, date.from, date.to, trendGroupBy)
            : Promise.resolve([]),
        ]);
        
        setReport(reportData);
        setExpenses(expensesData);
        setCategoryTotals(categoryData);
        setTrendData(trendDataRaw);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch report', err);
        setError('Failed to load report. The link may have expired.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [date, trendGroupBy]);

  const handlePresetSelect = (range: DateRange, preset: DatePreset) => {
    setDate(range);
    setCurrentPreset(preset);
  };

  if (loading && !report) {
    return (
      <div className="min-h-screen bg-background font-sans text-text flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary"></div>
          <p className="text-sm text-text/60">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background font-sans text-text flex flex-col items-center justify-center p-4">
        <div className="p-6 bg-surface rounded-xl border border-white/10 text-center max-w-md">
          <div className="w-12 h-12 bg-rose-500/10 rounded-full flex items-center justify-center mx-auto mb-4 text-rose-400">
            <CalendarIcon className="w-6 h-6" />
          </div>
          <h2 className="text-xl font-bold mb-2">Access Denied</h2>
          <p className="text-text/60">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background font-sans text-text selection:bg-primary/30">
      <Sidebar />
      <TopBar />

      <main className="pl-64 pt-20">
        <div className="p-8 max-w-[1800px] mx-auto space-y-6">
          
          {/* Header & Date Controls */}
          <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-4 mb-6">
            <div>
              <h1 className="text-3xl font-bold text-text tracking-tight mb-1" style={{ fontFamily: 'Fira Code, monospace' }}>
                Expense Dashboard
              </h1>
              <p className="text-text/60" style={{ fontFamily: 'Fira Sans, sans-serif' }}>
                Comprehensive view of your spending habits  
              </p>
            </div>
            <div className="flex flex-col gap-3">
              <DateRangePresets onSelectPreset={handlePresetSelect} currentPreset={currentPreset} />
              <DatePickerWithRange date={date} setDate={(range) => { setDate(range); setCurrentPreset('custom'); }} />
            </div>
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <DashboardCard className="relative overflow-hidden group cursor-pointer hover:border-primary/30 transition-all">
              <div className="relative z-10">
                <div className="flex justify-between items-start mb-3">
                  <div className="p-2.5 rounded-lg bg-primary/10 text-primary ring-1 ring-inset ring-primary/20">
                    <CurrencyDollarIcon className="w-5 h-5" />
                  </div>
                </div>
                <div className="space-y-1">
                  <p className="text-text/60 text-xs font-medium uppercase tracking-wider" style={{ fontFamily: 'Fira Sans, sans-serif' }}>
                    Total Expenses
                  </p>
                  <h3 className="text-2xl font-bold text-text tracking-tight" style={{ fontFamily: 'Fira Code, monospace' }}>
                    ${report?.total_expenses.toFixed(2)}
                  </h3>
                </div>
              </div>
            </DashboardCard>

            <DashboardCard className="relative overflow-hidden group cursor-pointer hover:border-primary/30 transition-all">
              <div className="relative z-10">
                <div className="flex justify-between items-start mb-3">
                  <div className="p-2.5 rounded-lg bg-primary/10 text-primary ring-1 ring-inset ring-primary/20">
                    <ListBulletIcon className="w-5 h-5" />
                  </div>
                </div>
                <div className="space-y-1">
                  <p className="text-text/60 text-xs font-medium uppercase tracking-wider" style={{ fontFamily: 'Fira Sans, sans-serif' }}>
                    Transactions
                  </p>
                  <h3 className="text-2xl font-bold text-text tracking-tight" style={{ fontFamily: 'Fira Code, monospace' }}>
                    {report?.transaction_count}
                  </h3>
                </div>
              </div>
            </DashboardCard>

            <DashboardCard className="relative overflow-hidden group cursor-pointer hover:border-primary/30 transition-all">
              <div className="relative z-10">
                <div className="flex justify-between items-start mb-3">
                  <div className="p-2.5 rounded-lg bg-orange-500/10 text-orange-500 ring-1 ring-inset ring-orange-500/20">
                    <ArrowTrendingUpIcon className="w-5 h-5" />
                  </div>
                </div>
                <div className="space-y-1">
                  <p className="text-text/60 text-xs font-medium uppercase tracking-wider" style={{ fontFamily: 'Fira Sans, sans-serif' }}>
                    Average / Tx
                  </p>
                  <h3 className="text-2xl font-bold text-text tracking-tight" style={{ fontFamily: 'Fira Code, monospace' }}>
                    ${report?.average_expense.toFixed(2)}
                  </h3>
                </div>
              </div>
            </DashboardCard>

            <DashboardCard className="relative overflow-hidden group cursor-pointer hover:border-primary/30 transition-all">
              <div className="relative z-10">
                <div className="flex justify-between items-start mb-3">
                  <div className="p-2.5 rounded-lg bg-primary/10 text-primary ring-1 ring-inset ring-primary/20">
                    <TagIcon className="w-5 h-5" />
                  </div>
                </div>
                <div className="space-y-1">
                  <p className="text-text/60 text-xs font-medium uppercase tracking-wider" style={{ fontFamily: 'Fira Sans, sans-serif' }}>
                    Top Category
                  </p>
                  <h3 className="text-lg font-bold text-text tracking-tight truncate" style={{ fontFamily: 'Fira Code, monospace' }}>
                    {categoryTotals[0]?.category_name || 'N/A'}
                  </h3>
                  <p className="text-xs text-text/50" style={{ fontFamily: 'Fira Sans, sans-serif' }}>
                    ${categoryTotals[0]?.total.toFixed(2) || '0.00'}
                  </p>
                </div>
              </div>
            </DashboardCard>
          </div>

          {/* Main Content - Two Column Layout */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            
            {/* Left Column - Expense List (60%) */}
            <div className="lg:col-span-7">
              <DashboardCard 
                title={
                  <div className="flex items-center gap-2">
                    <ListBulletIcon className="w-5 h-5" />
                    <span style={{ fontFamily: 'Fira Sans, sans-serif' }}>All Expenses</span>
                  </div>
                }
                className="h-[700px]"
              >
                <ExpenseList expenses={expenses} />
              </DashboardCard>
            </div>

            {/* Right Column - Charts (40%) */}
            <div className="lg:col-span-5 space-y-6">
              
              {/* Spending Trend Chart */}
              <DashboardCard 
                title={
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <ChartBarIcon className="w-5 h-5" />
                      <span style={{ fontFamily: 'Fira Sans, sans-serif' }}>Spending Trend</span>
                    </div>
                    <div className="flex gap-1">
                      {(['day', 'week', 'month'] as const).map((option) => (
                        <button
                          key={option}
                          onClick={() => setTrendGroupBy(option)}
                          className={`px-2 py-1 text-xs font-medium rounded transition-all cursor-pointer
                            ${trendGroupBy === option
                              ? 'bg-primary text-white'
                              : 'text-text/60 hover:text-text hover:bg-white/5'
                            }
                          `}
                          style={{ fontFamily: 'Fira Sans, sans-serif' }}
                        >
                          {option.charAt(0).toUpperCase() + option.slice(1)}
                        </button>
                      ))}
                    </div>
                  </div>
                }
                className="h-[350px]"
              >
                <SpendingTrendChart data={trendData} groupBy={trendGroupBy} className="h-[280px]" />
              </DashboardCard>

              {/* Category Breakdown */}
              <DashboardCard 
                title={
                  <div className="flex items-center gap-2">
                    <TagIcon className="w-5 h-5" />
                    <span style={{ fontFamily: 'Fira Sans, sans-serif' }}>Category Breakdown</span>
                  </div>
                }
                className="h-[320px]"
              >
                <div className="h-[250px] w-full flex items-center justify-center">
                  {categoryTotals && categoryTotals.length > 0 ? (
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={categoryTotals}
                          cx="50%"
                          cy="50%"
                          innerRadius={60}
                          outerRadius={90}
                          paddingAngle={3}
                          dataKey="total"
                          nameKey="category_name"
                        >
                          {categoryTotals.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} stroke="rgba(0,0,0,0.1)" strokeWidth={2} />
                          ))}
                        </Pie>
                        <Tooltip 
                          contentStyle={{ 
                            backgroundColor: '#1e293b', 
                            borderColor: '#334155', 
                            color: '#fff', 
                            borderRadius: '0.75rem',
                            fontFamily: 'Fira Sans, sans-serif'
                          }}
                          itemStyle={{ color: '#fff', fontFamily: 'Fira Code, monospace' }}
                          formatter={(value: number) => `$${value.toFixed(2)}`}
                        />
                        <Legend 
                          layout="vertical" 
                          verticalAlign="middle" 
                          align="right"
                          wrapperStyle={{ fontSize: '11px', color: '#94a3b8', fontFamily: 'Fira Sans, sans-serif' }}
                        />
                      </PieChart>
                    </ResponsiveContainer>
                  ) : (
                    <div className="text-text/40 text-sm" style={{ fontFamily: 'Fira Sans, sans-serif' }}>
                      No category data available
                    </div>
                  )}
                </div>
              </DashboardCard>
            </div>
          </div>
          
        </div>
      </main>
    </div>
  );
}
