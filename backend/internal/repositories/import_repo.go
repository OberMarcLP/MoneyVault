package repositories

import (
	"encoding/json"

	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ImportRepository struct {
	db *sqlx.DB
}

func NewImportRepository(db *sqlx.DB) *ImportRepository {
	return &ImportRepository{db: db}
}

func (r *ImportRepository) Create(job *models.ImportJob) error {
	mappingJSON, _ := json.Marshal(job.ColumnMapping)
	query := `INSERT INTO import_jobs (id, user_id, account_id, filename, status, total_rows, imported_rows, duplicate_rows, error_message, column_mapping)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(query, job.ID, job.UserID, job.AccountID, job.Filename, job.Status, job.TotalRows, job.ImportedRows, job.DuplicateRows, job.ErrorMessage, mappingJSON)
	return err
}

// CreateWithTx creates an import job within a database transaction.
func (r *ImportRepository) CreateWithTx(dbTx *sqlx.Tx, job *models.ImportJob) error {
	mappingJSON, _ := json.Marshal(job.ColumnMapping)
	query := `INSERT INTO import_jobs (id, user_id, account_id, filename, status, total_rows, imported_rows, duplicate_rows, error_message, column_mapping)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := dbTx.Exec(query, job.ID, job.UserID, job.AccountID, job.Filename, job.Status, job.TotalRows, job.ImportedRows, job.DuplicateRows, job.ErrorMessage, mappingJSON)
	return err
}

func (r *ImportRepository) List(userID uuid.UUID) ([]models.ImportJob, error) {
	rows, err := r.db.Query(`SELECT id, user_id, account_id, filename, status, total_rows, imported_rows,
		duplicate_rows, error_message, column_mapping, created_at, updated_at
		FROM import_jobs WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.ImportJob
	for rows.Next() {
		var job models.ImportJob
		var mappingBytes []byte
		if err := rows.Scan(&job.ID, &job.UserID, &job.AccountID, &job.Filename, &job.Status,
			&job.TotalRows, &job.ImportedRows, &job.DuplicateRows, &job.ErrorMessage,
			&mappingBytes, &job.CreatedAt, &job.UpdatedAt); err != nil {
			continue
		}
		_ = json.Unmarshal(mappingBytes, &job.ColumnMapping)
		jobs = append(jobs, job)
	}
	if jobs == nil {
		jobs = []models.ImportJob{}
	}
	return jobs, nil
}

func (r *ImportRepository) UpdateStatus(id uuid.UUID, status models.ImportStatus, imported, duplicates int, errMsg *string) error {
	query := `UPDATE import_jobs SET status = $1, imported_rows = $2, duplicate_rows = $3,
		error_message = $4, updated_at = NOW() WHERE id = $5`
	_, err := r.db.Exec(query, status, imported, duplicates, errMsg, id)
	return err
}
