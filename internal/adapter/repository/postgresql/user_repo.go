package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (user_id, messenger_type, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.UserID, user.MessengerType, user.CreatedAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	const query = `
		SELECT user_id, messenger_type, created_at
		FROM users
		WHERE user_id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID, &user.MessengerType, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE user_id = $1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
