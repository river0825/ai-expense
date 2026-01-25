package sqlite

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
		WHERE provider = ? AND model = ? AND is_active = 1 AND effective_date <= ?
		ORDER BY effective_date DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, provider, model, time.Now().UTC())

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
		WHERE is_active = 1
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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		SET input_token_price = ?, output_token_price = ?, is_active = ?, updated_at = ?
		WHERE id = ?
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
		SET is_active = 0, updated_at = ?
		WHERE provider = ? AND model = ?
	`
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), provider, model)
	return err
}
