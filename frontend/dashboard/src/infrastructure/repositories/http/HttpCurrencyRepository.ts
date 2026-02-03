import axios from 'axios';
import { CurrencyRepository } from '@/domain/repositories/CurrencyRepository';
import { Currency } from '@/domain/models/Currency';

export class HttpCurrencyRepository implements CurrencyRepository {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://api.aiexpense.net';
  }

  async getCurrencies(): Promise<Currency[]> {
    try {
      const url = `${this.baseURL}/api/currencies`;
      const response = await axios.get<{ status: string; data: Currency[] }>(url);
      return response.data.data;
    } catch (error) {
      console.error('Failed to fetch currencies:', error);
      throw error;
    }
  }
}
