package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifBudgetAlert    NotificationType = "budget_alert"
	NotifPriceAlert     NotificationType = "price_alert"
	NotifMilestone      NotificationType = "milestone"
	NotifInfo           NotificationType = "info"
	NotifImportComplete NotificationType = "import_complete"
	NotifSummary        NotificationType = "summary"
)

type Notification struct {
	ID        uuid.UUID        `db:"id" json:"id"`
	UserID    uuid.UUID        `db:"user_id" json:"user_id"`
	Type      NotificationType `db:"type" json:"type"`
	Title     string           `db:"title" json:"title"`
	Message   string           `db:"message" json:"message"`
	IsRead    bool             `db:"is_read" json:"is_read"`
	Link      string           `db:"link" json:"link"`
	CreatedAt time.Time        `db:"created_at" json:"created_at"`
}

type AlertRuleType string

const (
	AlertBudgetOverspend AlertRuleType = "budget_overspend"
	AlertPriceDrop       AlertRuleType = "price_drop"
	AlertPriceRise       AlertRuleType = "price_rise"
	AlertNetWorthMilestone AlertRuleType = "net_worth_milestone"
)

type AlertRule struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	UserID        uuid.UUID       `db:"user_id" json:"user_id"`
	Type          AlertRuleType   `db:"type" json:"type"`
	Condition     json.RawMessage `db:"condition" json:"condition"`
	IsActive      bool            `db:"is_active" json:"is_active"`
	LastTriggered *time.Time      `db:"last_triggered" json:"last_triggered"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
	DeletedAt     *time.Time      `db:"deleted_at" json:"-"`
}

type CreateAlertRuleRequest struct {
	Type      string          `json:"type" binding:"required"`
	Condition json.RawMessage `json:"condition" binding:"required"`
}

// Condition structs for different alert types
type BudgetAlertCondition struct {
	Threshold float64 `json:"threshold"` // percentage, e.g. 80 or 100
}

type PriceAlertCondition struct {
	Symbol    string  `json:"symbol"`
	Direction string  `json:"direction"` // "above" or "below"
	Price     float64 `json:"price"`
}

type NetWorthMilestoneCondition struct {
	Amount float64 `json:"amount"`
}
