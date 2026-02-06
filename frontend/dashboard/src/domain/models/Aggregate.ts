import { Category } from './Category';
import { Currency } from './Currency';
import { User } from './User';

export interface Account {
  name: string;
  created_at: string;
}

export interface AggregateSettings {
  currencies: Currency[];
  profile: User;
  categories: Category[];
  accounts: Account[];
}
