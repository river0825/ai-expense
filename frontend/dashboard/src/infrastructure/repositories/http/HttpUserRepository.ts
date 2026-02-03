import axios from 'axios';
import { UserRepository } from '@/domain/repositories/UserRepository';
import { User, UserSettings } from '@/domain/models/User';

export class HttpUserRepository implements UserRepository {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://aiexpense-996531141309.us-central1.run.app';
  }

  async getUser(token: string = 'test-user'): Promise<User> {
    try {
      // Assuming GET /api/user returns the current user profile
      const url = `${this.baseURL}/api/user?token=${token}`;
      const response = await axios.get<{ status: string; data: User }>(url);
      return response.data.data;
    } catch (error) {
      console.error('Failed to fetch user:', error);
      throw error;
    }
  }

  async updateSettings(token: string = 'test-user', settings: UserSettings): Promise<void> {
    try {
      // Assuming PUT /api/user/settings updates the settings
      // Note: Backend implementation plan said: GET/PUT /api/user/settings
      const url = `${this.baseURL}/api/user/settings?token=${token}`;
      await axios.put(url, settings);
    } catch (error) {
      console.error('Failed to update user settings:', error);
      throw error;
    }
  }
}
