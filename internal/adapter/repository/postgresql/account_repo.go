package postgresql

import (
	"database/sql"
	"errors"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ domain.AccountRepository = (*AccountRepository)(nil)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) GetByUserID(userID string) ([]*domain.Account, error) {
	const query = `
		SELECT user_id, name, created_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		var acc domain.Account
		if err := rows.Scan(&acc.UserID, &acc.Name, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, &acc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *AccountRepository) Create(account *domain.Account) error {
	const query = `
		INSERT INTO accounts (user_id, name, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(query, account.UserID, account.Name, account.CreatedAt)
	return err
}

func (r *AccountRepository) Update(userID string, oldName string, newName string) error {
	const query = `
		UPDATE accounts
		SET name = $1
		WHERE user_id = $2 AND name = $3
	`

	result, err := r.db.Exec(query, newName, userID, oldName)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("account not found")
	}

	return nil
}

func (r *AccountRepository) Delete(userID string, name string) error {
	const query = `
		DELETE FROM accounts
		WHERE user_id = $1 AND name = $2
	`

	result, err := r.db.Exec(query, userID, name)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("account not found")
	}

	return nil
}
