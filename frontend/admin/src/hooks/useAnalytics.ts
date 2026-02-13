import { useState, useEffect } from 'react';
import { AnalyticsData, AtRiskAccount, Period } from '../lib/types';
import { fetchAnalytics, fetchAtRiskAccounts } from '../lib/api';

interface UseAnalyticsResult {
  data: AnalyticsData | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

interface UseAtRiskResult {
  data: AtRiskAccount[];
  loading: boolean;
  error: string | null;
}

export const useAnalytics = (period: Period): UseAnalyticsResult => {
  const [data, setData] = useState<AnalyticsData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetchAnalytics(period);
      setData(response.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch analytics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [period]);

  return { data, loading, error, refetch: fetchData };
};

export const useAtRiskAccounts = (): UseAtRiskResult => {
  const [data, setData] = useState<AtRiskAccount[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const response = await fetchAtRiskAccounts();
        setData(response.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch at-risk accounts');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  return { data, loading, error };
};
