export interface Category {
  id: string;
  user_id: string;
  name: string;
  description: string;
  is_default: boolean;
}

export interface MergeResult {
  merged_count: number;
  message: string;
}
