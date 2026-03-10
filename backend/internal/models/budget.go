package models

import (
	"time"

	"github.com/google/uuid"
)

type BudgetPeriod string

const (
	PeriodWeekly  BudgetPeriod = "weekly"
	PeriodMonthly BudgetPeriod = "monthly"
	PeriodYearly  BudgetPeriod = "yearly"
)

type Budget struct {
	ID         uuid.UUID    `db:"id" json:"id"`
	UserID     uuid.UUID    `db:"user_id" json:"user_id"`
	CategoryID *uuid.UUID   `db:"category_id" json:"category_id"`
	Amount     float64      `db:"amount" json:"amount"`
	Period     BudgetPeriod `db:"period" json:"period"`
	StartDate  time.Time    `db:"start_date" json:"start_date"`
	EndDate    *time.Time   `db:"end_date" json:"end_date"`
	Rollover   bool         `db:"rollover" json:"rollover"`
	CreatedAt  time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time   `db:"deleted_at" json:"-"`
}

type BudgetWithSpending struct {
	Budget
	Spent        float64 `json:"spent"`
	Remaining    float64 `json:"remaining"`
	Percentage   float64 `json:"percentage"`
	CategoryName string  `json:"category_name"`
	CategoryIcon string  `json:"category_icon"`
	CategoryColor string `json:"category_color"`
}

type CreateBudgetRequest struct {
	CategoryID string  `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Period     string  `json:"period" binding:"required,oneof=weekly monthly yearly"`
	StartDate  string  `json:"start_date" binding:"required"`
	EndDate    string  `json:"end_date"`
	Rollover   bool    `json:"rollover"`
}

type UpdateBudgetRequest struct {
	Amount   *float64 `json:"amount"`
	Period   string   `json:"period" binding:"omitempty,oneof=weekly monthly yearly"`
	EndDate  string   `json:"end_date"`
	Rollover *bool    `json:"rollover"`
}
