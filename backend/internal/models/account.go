package models

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountChecking    AccountType = "checking"
	AccountSavings     AccountType = "savings"
	AccountCredit      AccountType = "credit"
	AccountInvestment  AccountType = "investment"
	AccountCryptoWallet AccountType = "crypto_wallet"
)

type Account struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	UserID      uuid.UUID   `db:"user_id" json:"user_id"`
	Name        string      `db:"name" json:"name"`
	Type        AccountType `db:"type" json:"type"`
	Currency    string      `db:"currency" json:"currency"`
	Balance     string      `db:"balance" json:"balance"`
	Institution *string     `db:"institution" json:"institution"`
	IsActive    bool        `db:"is_active" json:"is_active"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time  `db:"deleted_at" json:"-"`
}

type CreateAccountRequest struct {
	Name        string      `json:"name" binding:"required"`
	Type        AccountType `json:"type" binding:"required,oneof=checking savings credit investment crypto_wallet"`
	Currency    string      `json:"currency" binding:"required,len=3"`
	Balance     string      `json:"balance" binding:"required"`
	Institution *string     `json:"institution"`
}

type UpdateAccountRequest struct {
	Name        *string      `json:"name"`
	Type        *AccountType `json:"type"`
	Currency    *string      `json:"currency"`
	Balance     *string      `json:"balance"`
	Institution *string      `json:"institution"`
	IsActive    *bool        `json:"is_active"`
}
