package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.PolicyRepository = (*PolicyRepository)(nil)

// PolicyRepository implements domain.PolicyRepository using SQLite
type PolicyRepository struct {
	db *sql.DB
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(db *sql.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

// GetByKey retrieves a policy by its unique key
func (r *PolicyRepository) GetByKey(ctx context.Context, key string) (*domain.Policy, error) {
	const query = `
		SELECT id, key, title, content, version, created_at, updated_at
		FROM policies
		WHERE key = ?
	`
	policy := &domain.Policy{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&policy.ID,
		&policy.Key,
		&policy.Title,
		&policy.Content,
		&policy.Version,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Policy not found
		}
		return nil, err
	}
	return policy, nil
}
