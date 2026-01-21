import { DashboardStats } from '../models/Stats';

export interface StatsRepository {
  getDashboardStats(): Promise<DashboardStats>;
}
