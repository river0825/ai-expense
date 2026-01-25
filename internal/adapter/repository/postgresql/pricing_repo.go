package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.PricingRepository = (*PricingRepository)(nil)

type PricingRepository struct {
	db *sql.DB
}

func NewPricingRepository(db *sql.DB) *PricingRepository {
	return &PricingRepository{db: db}
}

// GetByProviderAndModel retrieves the most recent active pricing for a provider/model combination
func (r *PricingRepository) GetByProviderAndModel(ctx context.Context, provider, model string) (*domain.PricingConfig, error) {
	const query = `
		SELECT id, provider, model, input_token_price, output_token_price,
		       currency, effective_date, is_active, created_at, updated_at
		FROM ai_pricing_config
		WHERE provider = $1 AND model = $2 AND is_active = true AND effective_date <= NOW()::DATE
		ORDER BY effective_date DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, provider, model)

	config := &domain.PricingConfig{}
	err := row.Scan(
		&config.ID, &config.Provider, &config.Model, &config.InputTokenPrice, &config.OutputTokenPrice,
		&config.Currency, &config.EffectiveDate, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No pricing found, not an error
		}
		return nil, err
	}

	return config, nil
}

// GetAll retrieves all active pricing configurations
func (r *PricingRepository) GetAll(ctx context.Context) ([]*domain.PricingConfig, error) {
	const query = `
		SELECT id, provider, model, input_token_price, output_token_price,
		       currency, effective_date, is_active, created_at, updated_at
		FROM ai_pricing_config
		WHERE is_active = true
		ORDER BY provider, model, effective_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*domain.PricingConfig
	for rows.Next() {
		config := &domain.PricingConfig{}
		err := rows.Scan(
			&config.ID, &config.Provider, &config.Model, &config.InputTokenPrice, &config.OutputTokenPrice,
			&config.Currency, &config.EffectiveDate, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, rows.Err()
}

// Create creates a new pricing configuration
func (r *PricingRepository) Create(ctx context.Context, config *domain.PricingConfig) error {
	const query = `
		INSERT INTO ai_pricing_config (
			id, provider, model, input_token_price, output_token_price,
			currency, effective_date, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		config.ID, config.Provider, config.Model, config.InputTokenPrice, config.OutputTokenPrice,
		config.Currency, config.EffectiveDate, config.IsActive, config.CreatedAt, config.UpdatedAt,
	)
	return err
}

// Update updates an existing pricing configuration
func (r *PricingRepository) Update(ctx context.Context, config *domain.PricingConfig) error {
	const query = `
		UPDATE ai_pricing_config
		SET input_token_price = $1, output_token_price = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query,
		config.InputTokenPrice, config.OutputTokenPrice, config.IsActive, time.Now().UTC(), config.ID,
	)
	return err
}

// Deactivate marks a pricing configuration as inactive
func (r *PricingRepository) Deactivate(ctx context.Context, provider, model string) error {
	const query = `
		UPDATE ai_pricing_config
		SET is_active = false, updated_at = NOW()
		WHERE provider = $1 AND model = $2
	`
	_, err := r.db.ExecContext(ctx, query, provider, model)
	return err
}
