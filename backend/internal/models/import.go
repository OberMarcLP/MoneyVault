package models

import (
	"time"

	"github.com/google/uuid"
)

type ImportStatus string

const (
	ImportPending    ImportStatus = "pending"
	ImportProcessing ImportStatus = "processing"
	ImportCompleted  ImportStatus = "completed"
	ImportFailed     ImportStatus = "failed"
)

type ImportJob struct {
	ID            uuid.UUID         `db:"id" json:"id"`
	UserID        uuid.UUID         `db:"user_id" json:"user_id"`
	AccountID     uuid.UUID         `db:"account_id" json:"account_id"`
	Filename      string            `db:"filename" json:"filename"`
	Status        ImportStatus      `db:"status" json:"status"`
	TotalRows     int               `db:"total_rows" json:"total_rows"`
	ImportedRows  int               `db:"imported_rows" json:"imported_rows"`
	DuplicateRows int               `db:"duplicate_rows" json:"duplicate_rows"`
	ErrorMessage  *string           `db:"error_message" json:"error_message"`
	ColumnMapping map[string]string `db:"column_mapping" json:"column_mapping"`
	CreatedAt     time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time         `db:"updated_at" json:"updated_at"`
}

type ColumnMapping struct {
	Date        string `json:"date"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Merchant    string `json:"merchant"`
	Category    string `json:"category"`
	SubCategory string `json:"sub_category"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Currency    string `json:"currency"`
}

type CSVPreviewRow struct {
	Values map[string]string `json:"values"`
}

type CSVPreview struct {
	Headers []string        `json:"headers"`
	Rows    []CSVPreviewRow `json:"rows"`
	Total   int             `json:"total"`
}

type ImportRequest struct {
	AccountID string        `json:"account_id" binding:"required"`
	Mapping   ColumnMapping `json:"mapping" binding:"required"`
}
