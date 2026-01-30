CREATE TABLE IF NOT EXISTS currencies (
    code TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    aliases JSON NOT NULL DEFAULT '[]'::json,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS currency_translations (
    id SERIAL PRIMARY KEY,
    currency_code TEXT NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(currency_code, locale)
);

CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    provider TEXT NOT NULL,
    base_currency TEXT NOT NULL,
    target_currency TEXT NOT NULL,
    rate NUMERIC NOT NULL,
    rate_date DATE NOT NULL,
    fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, base_currency, target_currency, rate_date)
);

INSERT INTO currencies (code, symbol, aliases)
VALUES
    ('TWD', 'NT$', '["台幣", "元", "新台幣", "NT$", "TWD"]'::json),
    ('JPY', '¥', '["日幣", "日元", "円", "JPY", "yen"]'::json),
    ('USD', '$', '["美元", "美金", "USD", "dollar", "$"]'::json),
    ('EUR', '€', '["歐元", "EUR", "euro", "€"]'::json),
    ('CNY', '¥', '["人民幣", "人民币", "RMB", "CNY", "¥"]'::json)
ON CONFLICT (code)
DO UPDATE SET symbol = excluded.symbol,
              aliases = excluded.aliases,
              updated_at = CURRENT_TIMESTAMP;

INSERT INTO currency_translations (currency_code, locale, name)
VALUES
    ('TWD', 'en', 'New Taiwan Dollar'),
    ('TWD', 'zh-TW', '台幣'),
    ('TWD', 'zh-CN', '新台币'),
    ('JPY', 'en', 'Japanese Yen'),
    ('JPY', 'zh-TW', '日幣'),
    ('JPY', 'zh-CN', '日元'),
    ('JPY', 'ja', '円'),
    ('USD', 'en', 'US Dollar'),
    ('USD', 'zh-TW', '美金'),
    ('USD', 'zh-CN', '美元'),
    ('EUR', 'en', 'Euro'),
    ('EUR', 'zh-TW', '歐元'),
    ('EUR', 'zh-CN', '欧元'),
    ('CNY', 'en', 'Chinese Yuan'),
    ('CNY', 'zh-TW', '人民幣'),
    ('CNY', 'zh-CN', '人民币')
ON CONFLICT (currency_code, locale)
DO UPDATE SET name = excluded.name;
