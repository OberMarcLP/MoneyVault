package models

import (
	"time"

	"github.com/google/uuid"
)

type Dividend struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	HoldingID uuid.UUID  `db:"holding_id" json:"holding_id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	Amount    string     `db:"amount" json:"amount"`
	Currency  string     `db:"currency" json:"currency"`
	ExDate    time.Time  `db:"ex_date" json:"ex_date"`
	PayDate   *time.Time `db:"pay_date" json:"pay_date"`
	Notes     string     `db:"notes" json:"notes"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
}

type CreateDividendRequest struct {
	HoldingID string  `json:"holding_id" binding:"required"`
	Amount    string  `json:"amount" binding:"required"`
	Currency  string  `json:"currency"`
	ExDate    string  `json:"ex_date" binding:"required"`
	PayDate   string  `json:"pay_date"`
	Notes     string  `json:"notes"`
}

type DividendSummary struct {
	TotalDividends  float64 `json:"total_dividends"`
	DividendsYTD    float64 `json:"dividends_ytd"`
	DividendCount   int     `json:"dividend_count"`
	LastDividendAt  string  `json:"last_dividend_at,omitempty"`
}
