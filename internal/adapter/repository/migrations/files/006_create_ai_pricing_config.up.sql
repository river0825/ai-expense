-- Create ai_pricing_config table to store provider and model pricing information
CREATE TABLE IF NOT EXISTS ai_pricing_config (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  model TEXT NOT NULL,
  input_token_price DECIMAL(10, 10) NOT NULL,
  output_token_price DECIMAL(10, 10) NOT NULL,
  currency TEXT NOT NULL DEFAULT 'USD',
  effective_date DATE NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Unique constraint: one active pricing per provider/model combination at a time
CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_pricing_config_provider_model_date
  ON ai_pricing_config(provider, model, effective_date);

-- Index for lookup by provider and model with active status
CREATE INDEX IF NOT EXISTS idx_ai_pricing_config_provider_model_active
  ON ai_pricing_config(provider, model, is_active);

-- Index for finding active pricing by effective_date
CREATE INDEX IF NOT EXISTS idx_ai_pricing_config_effective_date
  ON ai_pricing_config(effective_date, is_active);

-- Seed initial Gemini pricing (USD per 1M tokens)
-- Source: Google Gemini Pricing as of January 2025
INSERT INTO ai_pricing_config (id, provider, model, input_token_price, output_token_price, currency, effective_date, is_active, created_at, updated_at)
VALUES (
  'pricing_gemini_2.5_lite_2025_01_25',
  'gemini',
  'gemini-2.5-lite',
  0.000000075,  -- $0.075 per 1M input tokens
  0.0000003,    -- $0.3 per 1M output tokens
  'USD',
  CURRENT_DATE,
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO NOTHING;
