package domain

import "time"

// AggregateSettings contains all aggregated user data for the API response
type AggregateSettings struct {
	Profile    *User       `json:"profile"`
	Categories []*Category `json:"categories"`
	Accounts   []*Account  `json:"accounts"`
	Currencies []Currency  `json:"currencies"`
}

// Account represents a user's account (like bank account or wallet)
type Account struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// AccountRepository defines the interface for account storage
type AccountRepository interface {
	GetByUserID(userID string) ([]*Account, error)
	Create(account *Account) error
	Update(userID string, oldName string, newName string) error
	Delete(userID string, name string) error
}
