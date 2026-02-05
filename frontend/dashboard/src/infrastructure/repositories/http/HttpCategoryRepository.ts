import axios from 'axios';
import { CategoryRepository } from '@/domain/repositories/CategoryRepository';
import { Category, MergeResult } from '@/domain/models/Category';

export class HttpCategoryRepository implements CategoryRepository {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'https://aiexpense-996531141309.us-central1.run.app';
  }

  async list(token: string): Promise<Category[]> {
    try {
      const response = await axios.get(`${this.baseURL}/api/user/categories`, {
        params: { user_id: token }
      });
      
      if (response.data.status === 'success') {
        return response.data.data || [];
      }
      throw new Error(response.data.error || 'Failed to list categories');
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error || error.message);
      }
      throw error;
    }
  }

  async create(token: string, name: string, description?: string): Promise<Category> {
    try {
      const response = await axios.post(`${this.baseURL}/api/user/categories`, {
        user_id: token,
        name,
        description: description || '',
        keywords: []
      });
      
      if (response.data.status === 'success') {
        return response.data.data;
      }
      throw new Error(response.data.error || 'Failed to create category');
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error || error.message);
      }
      throw error;
    }
  }

  async update(token: string, id: string, name: string, description?: string): Promise<Category> {
    try {
      const response = await axios.put(`${this.baseURL}/api/user/categories`, {
        id,
        user_id: token,
        name,
        description: description || '',
        keywords: []
      });
      
      if (response.data.status === 'success') {
        return response.data.data;
      }
      throw new Error(response.data.error || 'Failed to update category');
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error || error.message);
      }
      throw error;
    }
  }

  async delete(token: string, id: string): Promise<void> {
    try {
      const response = await axios.delete(`${this.baseURL}/api/user/categories`, {
        data: { id, user_id: token }
      });
      
      if (response.data.status !== 'success') {
        throw new Error(response.data.error || 'Failed to delete category');
      }
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error || error.message);
      }
      throw error;
    }
  }

  async merge(token: string, sourceId: string, targetId: string): Promise<MergeResult> {
    try {
      const response = await axios.post(`${this.baseURL}/api/user/categories/merge`, {
        user_id: token,
        source_id: sourceId,
        target_id: targetId
      });
      
      if (response.data.status === 'success') {
        return response.data.data;
      }
      throw new Error(response.data.error || 'Failed to merge categories');
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error || error.message);
      }
      throw error;
    }
  }
}
