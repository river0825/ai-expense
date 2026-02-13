import { AnalyticsResponse, AtRiskResponse, Metric, Period } from './types';
import { getToken } from './auth';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/admin';

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

async function fetchWithAuth<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    // Handle unauthorized - maybe redirect to login
    if (typeof window !== 'undefined') {
      window.location.href = '/login';
    }
    throw new ApiError(401, 'Unauthorized');
  }

  if (!response.ok) {
    throw new ApiError(response.status, 'API request failed');
  }

  const data = await response.json();
  return data;
}

export const fetchAnalytics = async (period: Period): Promise<AnalyticsResponse> => {
  const compareMap: Record<Period, string> = {
    '7d': 'prev_7d',
    '30d': 'prev_30d',
    '90d': 'prev_90d',
  };
  
  const raw = await fetchWithAuth<any>(`/analytics/overview?period=${period}&compare=${compareMap[period]}`);

  if (raw?.data?.metrics) {
    return raw as AnalyticsResponse;
  }

  const now = new Date();
  const startDate = new Date(now);
  if (period === '7d') startDate.setDate(now.getDate() - 7);
  if (period === '30d') startDate.setDate(now.getDate() - 30);
  if (period === '90d') startDate.setDate(now.getDate() - 90);

  const metric = (current: number): Metric => ({ current, previous: 0, delta_percent: 0 });

  const totalExpenses = Number(raw?.data?.expenses?.total_expenses ?? 0);
  const averageExpensePerUser = Number(raw?.data?.growth?.average_expense_per_user ?? 0);
  const weeklyGrowth = Number(raw?.data?.growth?.weekly_growth_percent ?? 0);
  const dailyGrowth = Number(raw?.data?.growth?.daily_growth_percent ?? 0);

  return {
    status: raw?.status ?? 'success',
    data: {
      period,
      start_date: startDate.toISOString(),
      end_date: now.toISOString(),
      metrics: {
        mrr: metric(totalExpenses),
        nrr: metric(weeklyGrowth),
        grr: metric(dailyGrowth),
        churn_rate: metric(Math.max(0, 100 - averageExpensePerUser)),
      },
    },
  };
};

// Mock implementation for now as endpoint is not defined in prompt
export const fetchAtRiskAccounts = async (): Promise<AtRiskResponse> => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 500));
  
  return {
    status: 'success',
    data: [
      {
        id: '1',
        name: 'Acme Corp',
        mrr: 5000,
        health_score: 45,
        last_activity: '2025-02-10T10:00:00Z',
        risk_factor: 'Low Usage',
      },
      {
        id: '2',
        name: 'Globex Inc',
        mrr: 12000,
        health_score: 30,
        last_activity: '2025-02-01T14:30:00Z',
        risk_factor: 'Support Tickets',
      },
      {
        id: '3',
        name: 'Soylent Corp',
        mrr: 8500,
        health_score: 55,
        last_activity: '2025-02-11T09:15:00Z',
        risk_factor: 'Payment Failed',
      },
      {
        id: '4',
        name: 'Initech',
        mrr: 3200,
        health_score: 20,
        last_activity: '2025-01-25T16:45:00Z',
        risk_factor: 'Churn Risk',
      },
      {
        id: '5',
        name: 'Umbrella Corp',
        mrr: 15000,
        health_score: 60,
        last_activity: '2025-02-12T11:20:00Z',
        risk_factor: 'Contract Expiry',
      },
    ],
  };
};
