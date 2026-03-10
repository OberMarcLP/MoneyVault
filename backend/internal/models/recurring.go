package models

import (
	"time"

	"github.com/google/uuid"
)

type Frequency string

const (
	FreqDaily     Frequency = "daily"
	FreqWeekly    Frequency = "weekly"
	FreqBiweekly  Frequency = "biweekly"
	FreqMonthly   Frequency = "monthly"
	FreqQuarterly Frequency = "quarterly"
	FreqYearly    Frequency = "yearly"
)

type RecurringTransaction struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	UserID            uuid.UUID  `db:"user_id" json:"user_id"`
	AccountID         uuid.UUID  `db:"account_id" json:"account_id"`
	Type              string     `db:"type" json:"type"`
	Amount            string     `db:"amount" json:"amount"`
	Currency          string     `db:"currency" json:"currency"`
	CategoryID        *uuid.UUID `db:"category_id" json:"category_id"`
	Description       string     `db:"description" json:"description"`
	Frequency         Frequency  `db:"frequency" json:"frequency"`
	NextDate          time.Time  `db:"next_date" json:"next_date"`
	EndDate           *time.Time `db:"end_date" json:"end_date"`
	TransferAccountID *uuid.UUID `db:"transfer_account_id" json:"transfer_account_id"`
	IsActive          bool       `db:"is_active" json:"is_active"`
	LastCreated       *time.Time `db:"last_created" json:"last_created"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt         *time.Time `db:"deleted_at" json:"-"`
}

type CreateRecurringRequest struct {
	AccountID         string `json:"account_id" binding:"required"`
	Type              string `json:"type" binding:"required,oneof=income expense transfer"`
	Amount            string `json:"amount" binding:"required"`
	Currency          string `json:"currency"`
	CategoryID        string `json:"category_id"`
	Description       string `json:"description"`
	Frequency         string `json:"frequency" binding:"required,oneof=daily weekly biweekly monthly quarterly yearly"`
	NextDate          string `json:"next_date" binding:"required"`
	EndDate           string `json:"end_date"`
	TransferAccountID string `json:"transfer_account_id"`
}
