import axios from 'axios';
import { format } from 'date-fns';
import { ReportRepository } from '@/domain/repositories/ReportRepository';
import { ExpenseReport } from '@/domain/models/Report';

export class HttpReportRepository implements ReportRepository {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://aiexpense-996531141309.us-central1.run.app';
  }

  async getReportSummary(token: string, startDate?: Date, endDate?: Date): Promise<ExpenseReport> {
    try {
      let url = `${this.baseURL}/api/reports/summary?token=${token}`;
      if (startDate) {
        url += `&start_date=${format(startDate, 'yyyy-MM-dd')}`;
      }
      if (endDate) {
        url += `&end_date=${format(endDate, 'yyyy-MM-dd')}`;
      }

      const response = await axios.get<{ status: string; data: ExpenseReport }>(url);
      return response.data.data;
    } catch (error) {
      console.error('Failed to fetch report summary:', error);
      throw error;
    }
  }
}
