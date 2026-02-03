import { Currency } from '../models/Currency';

export interface CurrencyRepository {
  getCurrencies(): Promise<Currency[]>;
}
