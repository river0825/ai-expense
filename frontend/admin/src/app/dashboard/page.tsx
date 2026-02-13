'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAnalytics, useAtRiskAccounts } from '@/hooks/useAnalytics';
import { Period } from '@/lib/types';
import { KPIGrid } from '@/components/KPIGrid';
import { KPICard } from '@/components/KPICard';
import { TrendChart } from '@/components/TrendChart';
import { PeriodFilter } from '@/components/PeriodFilter';
import { AtRiskTable } from '@/components/AtRiskTable';
import { isAuthenticated } from '@/lib/auth';

export default function DashboardPage() {
  const router = useRouter();
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);
  const [period, setPeriod] = useState<Period>('30d');
  const { data: analytics, loading: analyticsLoading, error: analyticsError } = useAnalytics(period);
  const { data: atRisk, loading: atRiskLoading, error: atRiskError } = useAtRiskAccounts();

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push('/login');
    } else {
      setIsCheckingAuth(false);
    }
  }, [router]);

  if (isCheckingAuth) {
    return (
      <div className="flex items-center justify-center h-screen bg-background text-foreground">
        <div className="animate-pulse">Checking authentication...</div>
      </div>
    );
  }

  if (analyticsLoading || atRiskLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-background text-foreground">
        <div className="animate-pulse">Loading Command Center...</div>
      </div>
    );
  }

  if (analyticsError || atRiskError) {
    return (
      <div className="flex items-center justify-center h-screen bg-background text-red-500">
        Error: {analyticsError || atRiskError}
      </div>
    );
  }

  if (!analytics) return null;

  return (
    <div className="p-8 space-y-8 bg-background min-h-screen text-foreground font-sans">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Revenue Command Center</h1>
          <p className="text-muted-foreground mt-1">
            Overview for {new Date(analytics.start_date).toLocaleDateString()} - {new Date(analytics.end_date).toLocaleDateString()}
          </p>
        </div>
        <PeriodFilter value={period} onChange={setPeriod} />
      </div>

      <KPIGrid>
        <KPICard title="MRR" metric={analytics.metrics.mrr} prefix="$" />
        <KPICard title="NRR" metric={analytics.metrics.nrr} suffix="%" />
        <KPICard title="GRR" metric={analytics.metrics.grr} suffix="%" />
        <KPICard title="Churn Rate" metric={analytics.metrics.churn_rate} suffix="%" inverse />
      </KPIGrid>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-card border border-border rounded-lg p-6 shadow-sm">
            <h3 className="text-lg font-medium mb-4">Revenue Trend</h3>
            <TrendChart className="w-full text-primary" height={300} />
          </div>
          
          <div className="bg-card border border-border rounded-lg p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium">At-Risk Accounts</h3>
              <span className="text-xs bg-red-500/10 text-red-500 px-2 py-1 rounded-full font-medium">
                {atRisk.length} Attention Needed
              </span>
            </div>
            <AtRiskTable accounts={atRisk} />
          </div>
        </div>

        <div className="space-y-6">
           <div className="bg-card border border-border rounded-lg p-6 h-full shadow-sm">
             <h3 className="text-lg font-medium mb-4">Insights</h3>
             <p className="text-sm text-muted-foreground mb-6">
               Churn rate has decreased by {Math.abs(analytics.metrics.churn_rate.delta_percent).toFixed(2)}% compared to the previous period. 
               This is likely due to the new onboarding flow.
             </p>
             
             <div className="space-y-4">
               <div className="text-sm font-medium">Top Churn Reasons</div>
               
               <div className="space-y-2">
                 <div className="flex justify-between text-xs text-muted-foreground">
                   <span>Pricing</span>
                   <span>45%</span>
                 </div>
                 <div className="w-full bg-muted rounded-full h-1.5 overflow-hidden">
                   <div className="bg-red-500 h-1.5 rounded-full" style={{ width: '45%' }}></div>
                 </div>
               </div>
               
               <div className="space-y-2">
                 <div className="flex justify-between text-xs text-muted-foreground">
                   <span>Missing Features</span>
                   <span>30%</span>
                 </div>
                 <div className="w-full bg-muted rounded-full h-1.5 overflow-hidden">
                   <div className="bg-yellow-500 h-1.5 rounded-full" style={{ width: '30%' }}></div>
                 </div>
               </div>

               <div className="space-y-2">
                 <div className="flex justify-between text-xs text-muted-foreground">
                   <span>Support Quality</span>
                   <span>15%</span>
                 </div>
                 <div className="w-full bg-muted rounded-full h-1.5 overflow-hidden">
                   <div className="bg-blue-500 h-1.5 rounded-full" style={{ width: '15%' }}></div>
                 </div>
               </div>
             </div>

             <div className="mt-8 pt-6 border-t border-border">
                <h4 className="text-sm font-medium mb-3">Recommended Actions</h4>
                <ul className="space-y-2 text-xs text-muted-foreground">
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">•</span>
                    Schedule QBR with Acme Corp
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">•</span>
                    Review pricing tier for Globex Inc
                  </li>
                </ul>
             </div>
           </div>
        </div>
      </div>
    </div>
  );
}
