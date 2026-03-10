package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionIncome   TransactionType = "income"
	TransactionExpense  TransactionType = "expense"
	TransactionTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID                uuid.UUID       `db:"id" json:"id"`
	AccountID         uuid.UUID       `db:"account_id" json:"account_id"`
	UserID            uuid.UUID       `db:"user_id" json:"user_id"`
	Type              TransactionType `db:"type" json:"type"`
	Amount            string          `db:"amount" json:"amount"`
	Currency          string          `db:"currency" json:"currency"`
	CategoryID        *uuid.UUID      `db:"category_id" json:"category_id"`
	Description       string          `db:"description" json:"description"`
	Date              time.Time       `db:"date" json:"date"`
	Tags              json.RawMessage `db:"tags" json:"tags"`
	ImportSource      string          `db:"import_source" json:"import_source"`
	TransferAccountID *uuid.UUID      `db:"transfer_account_id" json:"transfer_account_id,omitempty"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
	DeletedAt         *time.Time      `db:"deleted_at" json:"-"`
}

type CreateTransactionRequest struct {
	AccountID         uuid.UUID       `json:"account_id" binding:"required"`
	Type              TransactionType `json:"type" binding:"required,oneof=income expense transfer"`
	Amount            string          `json:"amount" binding:"required"`
	Currency          string          `json:"currency" binding:"required,len=3"`
	CategoryID        *uuid.UUID      `json:"category_id"`
	Description       string          `json:"description"`
	Date              string          `json:"date" binding:"required"`
	Tags              json.RawMessage `json:"tags"`
	TransferAccountID *uuid.UUID      `json:"transfer_account_id"`
}

type UpdateTransactionRequest struct {
	AccountID         *uuid.UUID       `json:"account_id"`
	Type              *TransactionType `json:"type"`
	Amount            *string          `json:"amount"`
	Currency          *string          `json:"currency"`
	CategoryID        *uuid.UUID       `json:"category_id"`
	Description       *string          `json:"description"`
	Date              *string          `json:"date"`
	Tags              json.RawMessage  `json:"tags"`
	TransferAccountID *uuid.UUID       `json:"transfer_account_id"`
}

type TransactionFilter struct {
	AccountID  *uuid.UUID       `form:"account_id"`
	Type       *TransactionType `form:"type"`
	CategoryID *uuid.UUID       `form:"category_id"`
	DateFrom   *string          `form:"date_from"`
	DateTo     *string          `form:"date_to"`
	Search     *string          `form:"search"`
	Page       int              `form:"page,default=1"`
	PerPage    int              `form:"per_page,default=50"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}
