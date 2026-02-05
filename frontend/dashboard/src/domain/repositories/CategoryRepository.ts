import { Category, MergeResult } from '../models/Category';

export interface CategoryRepository {
  /**
   * List all categories for the current user
   */
  list(token: string): Promise<Category[]>;
  
  /**
   * Create a new category
   */
  create(token: string, name: string, description?: string): Promise<Category>;
  
  /**
   * Update an existing category
   */
  update(token: string, id: string, name: string, description?: string): Promise<Category>;
  
  /**
   * Delete a category
   */
  delete(token: string, id: string): Promise<void>;
  
  /**
   * Merge two categories
   */
  merge(token: string, sourceId: string, targetId: string): Promise<MergeResult>;
}
