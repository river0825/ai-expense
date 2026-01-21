package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

type PolicyRepository struct {
	db *sql.DB
}

func NewPolicyRepository(db *sql.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

func (r *PolicyRepository) GetByKey(ctx context.Context, key string) (*domain.Policy, error) {
	const query = `
		SELECT id, key, title, content, version, created_at, updated_at
		FROM policies
		WHERE key = $1
	`

	policy := &domain.Policy{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&policy.ID, &policy.Key, &policy.Title,
		&policy.Content, &policy.Version, &policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return policy, nil
}
