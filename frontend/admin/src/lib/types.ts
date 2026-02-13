export interface Metric {
  current: number;
  previous: number;
  delta_percent: number;
}

export interface AnalyticsData {
  period: string;
  start_date: string;
  end_date: string;
  metrics: {
    mrr: Metric;
    nrr: Metric;
    grr: Metric;
    churn_rate: Metric;
  };
}

export interface AnalyticsResponse {
  status: string;
  data: AnalyticsData;
}

export interface AtRiskAccount {
  id: string;
  name: string;
  mrr: number;
  health_score: number;
  last_activity: string;
  risk_factor: string;
}

export interface AtRiskResponse {
  status: string;
  data: AtRiskAccount[];
}

export type Period = '7d' | '30d' | '90d';
