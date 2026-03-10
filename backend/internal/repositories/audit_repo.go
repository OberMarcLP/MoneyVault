package repositories

import (
	"fmt"
	"moneyvault/internal/models"

	"github.com/jmoiron/sqlx"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(entry *models.AuditLog) error {
	query := `INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, ip_address, user_agent, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`
	_, err := r.db.Exec(query,
		entry.ID, entry.UserID, entry.Action,
		entry.ResourceType, entry.ResourceID,
		entry.IPAddress, entry.UserAgent, entry.Details,
	)
	return err
}

func (r *AuditRepository) List(filter models.AuditFilter) ([]models.AuditLog, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}

	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.UserID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.Action != "" {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, filter.Action)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM audit_logs " + where
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	query := fmt.Sprintf("SELECT * FROM audit_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1)
	args = append(args, filter.Limit, offset)

	var logs []models.AuditLog
	if err := r.db.Select(&logs, query, args...); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
