import { ExpenseReport } from '../models/Report';

export interface ReportRepository {
  getReportSummary(token: string, startDate?: Date, endDate?: Date): Promise<ExpenseReport>;
}
