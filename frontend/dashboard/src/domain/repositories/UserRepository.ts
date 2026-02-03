import { User, UserSettings } from '../models/User';

export interface UserRepository {
  getUser(token: string): Promise<User>;
  updateSettings(token: string, settings: UserSettings): Promise<void>;
}
