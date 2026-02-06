import axios from 'axios';
import { AggregateSettings } from '@/domain/models/Aggregate';

export class HttpAggregateRepository {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://aiexpense-996531141309.us-central1.run.app';
  }

  async getAggregate(token: string): Promise<AggregateSettings> {
    try {
      const url = `${this.baseURL}/api/aggregate?token=${token}`;
      const response = await axios.get<{ status: string; data: AggregateSettings }>(url);
      return response.data.data;
    } catch (error) {
      console.error('Failed to fetch aggregate data:', error);
      throw error;
    }
  }
}
