package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           uuid.UUID        `db:"id" json:"id"`
	UserID       *uuid.UUID       `db:"user_id" json:"user_id"`
	Action       string           `db:"action" json:"action"`
	ResourceType *string          `db:"resource_type" json:"resource_type"`
	ResourceID   *uuid.UUID       `db:"resource_id" json:"resource_id"`
	IPAddress    *string          `db:"ip_address" json:"ip_address"`
	UserAgent    *string          `db:"user_agent" json:"user_agent"`
	Details      *json.RawMessage `db:"details" json:"details"`
	CreatedAt    time.Time        `db:"created_at" json:"created_at"`
}

type AuditFilter struct {
	UserID *uuid.UUID `form:"user_id"`
	Action string     `form:"action"`
	Page   int        `form:"page"`
	Limit  int        `form:"limit"`
}
