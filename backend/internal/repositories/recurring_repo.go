package repositories

import (
	"time"

	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RecurringRepository struct {
	db *sqlx.DB
}

func NewRecurringRepository(db *sqlx.DB) *RecurringRepository {
	return &RecurringRepository{db: db}
}

func (r *RecurringRepository) Create(rt *models.RecurringTransaction) error {
	query := `INSERT INTO recurring_transactions (id, user_id, account_id, type, amount, currency, category_id,
		description, frequency, next_date, end_date, transfer_account_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.Exec(query, rt.ID, rt.UserID, rt.AccountID, rt.Type, rt.Amount, rt.Currency,
		rt.CategoryID, rt.Description, rt.Frequency, rt.NextDate, rt.EndDate, rt.TransferAccountID, rt.IsActive)
	return err
}

func (r *RecurringRepository) GetByID(userID, id uuid.UUID) (*models.RecurringTransaction, error) {
	var rt models.RecurringTransaction
	err := r.db.Get(&rt, `SELECT * FROM recurring_transactions WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return &rt, err
}

func (r *RecurringRepository) List(userID uuid.UUID) ([]models.RecurringTransaction, error) {
	var list []models.RecurringTransaction
	err := r.db.Select(&list, `SELECT * FROM recurring_transactions WHERE user_id = $1 AND deleted_at IS NULL ORDER BY next_date ASC`, userID)
	if list == nil {
		list = []models.RecurringTransaction{}
	}
	return list, err
}

func (r *RecurringRepository) GetDue(before time.Time) ([]models.RecurringTransaction, error) {
	var list []models.RecurringTransaction
	err := r.db.Select(&list, `SELECT * FROM recurring_transactions WHERE is_active = true AND next_date <= $1 AND deleted_at IS NULL`, before)
	if list == nil {
		list = []models.RecurringTransaction{}
	}
	return list, err
}

func (r *RecurringRepository) Update(rt *models.RecurringTransaction) error {
	query := `UPDATE recurring_transactions SET amount = $1, currency = $2, category_id = $3,
		description = $4, frequency = $5, next_date = $6, end_date = $7,
		transfer_account_id = $8, is_active = $9, last_created = $10, updated_at = NOW()
		WHERE id = $11 AND user_id = $12 AND deleted_at IS NULL`
	_, err := r.db.Exec(query, rt.Amount, rt.Currency, rt.CategoryID, rt.Description,
		rt.Frequency, rt.NextDate, rt.EndDate, rt.TransferAccountID, rt.IsActive, rt.LastCreated, rt.ID, rt.UserID)
	return err
}

func (r *RecurringRepository) Delete(userID, id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE recurring_transactions SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return err
}
