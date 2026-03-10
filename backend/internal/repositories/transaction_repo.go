package repositories

import (
	"fmt"
	"moneyvault/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) DB() *sqlx.DB { return r.db }

func (r *TransactionRepository) Create(tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, account_id, user_id, type, amount, currency, category_id, description, date, tags, import_source, transfer_account_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING created_at`
	return r.db.QueryRow(query,
		tx.ID, tx.AccountID, tx.UserID, tx.Type, tx.Amount,
		tx.Currency, tx.CategoryID, tx.Description, tx.Date,
		tx.Tags, tx.ImportSource, tx.TransferAccountID,
	).Scan(&tx.CreatedAt)
}

// CreateWithTx creates a transaction record within a database transaction.
func (r *TransactionRepository) CreateWithTx(dbTx *sqlx.Tx, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, account_id, user_id, type, amount, currency, category_id, description, date, tags, import_source, transfer_account_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING created_at`
	return dbTx.QueryRow(query,
		tx.ID, tx.AccountID, tx.UserID, tx.Type, tx.Amount,
		tx.Currency, tx.CategoryID, tx.Description, tx.Date,
		tx.Tags, tx.ImportSource, tx.TransferAccountID,
	).Scan(&tx.CreatedAt)
}

func (r *TransactionRepository) GetByID(id, userID uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.Get(&tx, "SELECT * FROM transactions WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL", id, userID)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepository) List(userID uuid.UUID, filter models.TransactionFilter) ([]models.Transaction, int, error) {
	conditions := []string{"user_id = $1", "deleted_at IS NULL"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.AccountID != nil {
		conditions = append(conditions, fmt.Sprintf("account_id = $%d", argIdx))
		args = append(args, *filter.AccountID)
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM transactions WHERE %s", where)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	query := fmt.Sprintf(
		"SELECT * FROM transactions WHERE %s ORDER BY date DESC, created_at DESC LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	var transactions []models.Transaction
	if err := r.db.Select(&transactions, query, args...); err != nil {
		return nil, 0, err
	}
	return transactions, total, nil
}

func (r *TransactionRepository) Update(tx *models.Transaction) error {
	query := `
		UPDATE transactions SET account_id = $3, type = $4, amount = $5,
		currency = $6, category_id = $7, description = $8, date = $9,
		tags = $10, transfer_account_id = $11
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(query,
		tx.ID, tx.UserID, tx.AccountID, tx.Type, tx.Amount,
		tx.Currency, tx.CategoryID, tx.Description, tx.Date,
		tx.Tags, tx.TransferAccountID,
	)
	return err
}

func (r *TransactionRepository) Delete(id, userID uuid.UUID) error {
	_, err := r.db.Exec("UPDATE transactions SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL", id, userID)
	return err
}

// ListExpensesByCategory fetches raw (encrypted) expense transactions for a category in a date range.
func (r *TransactionRepository) ListExpensesByCategory(userID uuid.UUID, categoryID *uuid.UUID, start, end time.Time) ([]models.Transaction, error) {
	var txs []models.Transaction
	err := r.db.Select(&txs,
		`SELECT * FROM transactions WHERE user_id = $1 AND category_id = $2 AND type = 'expense' AND date >= $3::date AND date < $4::date AND deleted_at IS NULL`,
		userID, categoryID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if txs == nil {
		txs = []models.Transaction{}
	}
	return txs, err
}

// ListByDateRange fetches raw (encrypted) transactions in a date range.
func (r *TransactionRepository) ListByDateRange(userID uuid.UUID, from, to time.Time) ([]models.Transaction, error) {
	var txs []models.Transaction
	err := r.db.Select(&txs,
		`SELECT * FROM transactions WHERE user_id = $1 AND date >= $2::date AND date < $3::date AND deleted_at IS NULL ORDER BY date DESC`,
		userID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if txs == nil {
		txs = []models.Transaction{}
	}
	return txs, err
}

// ListExpensesByDateRange fetches raw (encrypted) expense transactions in a date range.
func (r *TransactionRepository) ListExpensesByDateRange(userID uuid.UUID, from, to time.Time) ([]models.Transaction, error) {
	var txs []models.Transaction
	err := r.db.Select(&txs,
		`SELECT * FROM transactions WHERE user_id = $1 AND type = 'expense' AND date >= $2::date AND date < $3::date AND deleted_at IS NULL ORDER BY date DESC`,
		userID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if txs == nil {
		txs = []models.Transaction{}
	}
	return txs, err
}
