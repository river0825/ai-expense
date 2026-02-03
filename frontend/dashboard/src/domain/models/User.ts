export interface User {
  user_id: string;
  messenger_type: string;
  created_at: string;
  home_currency: string;
  locale: string;
}

export interface UserSettings {
  home_currency: string;
  locale: string;
}
