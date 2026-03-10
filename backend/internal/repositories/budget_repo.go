package repositories

import (
	"time"

	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BudgetRepository struct {
	db *sqlx.DB
}

func NewBudgetRepository(db *sqlx.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(b *models.Budget) error {
	query := `INSERT INTO budgets (id, user_id, category_id, amount, period, start_date, end_date, rollover)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(query, b.ID, b.UserID, b.CategoryID, b.Amount, b.Period, b.StartDate, b.EndDate, b.Rollover)
	return err
}

func (r *BudgetRepository) GetByID(userID, budgetID uuid.UUID) (*models.Budget, error) {
	var b models.Budget
	err := r.db.Get(&b, `SELECT * FROM budgets WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, budgetID, userID)
	return &b, err
}

func (r *BudgetRepository) List(userID uuid.UUID) ([]models.Budget, error) {
	var budgets []models.Budget
	err := r.db.Select(&budgets, `SELECT * FROM budgets WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`, userID)
	if budgets == nil {
		budgets = []models.Budget{}
	}
	return budgets, err
}

func (r *BudgetRepository) Update(b *models.Budget) error {
	query := `UPDATE budgets SET amount = $1, period = $2, end_date = $3, rollover = $4, updated_at = NOW()
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL`
	_, err := r.db.Exec(query, b.Amount, b.Period, b.EndDate, b.Rollover, b.ID, b.UserID)
	return err
}

func (r *BudgetRepository) Delete(userID, budgetID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE budgets SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, budgetID, userID)
	return err
}

func (r *BudgetRepository) GetSpentForCategory(userID uuid.UUID, categoryID *uuid.UUID, periodStart, periodEnd time.Time) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(CAST(amount AS NUMERIC)), 0) FROM transactions
		WHERE user_id = $1 AND category_id = $2 AND type = 'expense'
		AND date >= $3 AND date < $4 AND deleted_at IS NULL`
	err := r.db.Get(&total, query, userID, categoryID, periodStart, periodEnd)
	return total, err
}

func (r *BudgetRepository) GetSpentByCategories(userID uuid.UUID, categoryIDs []uuid.UUID, periodStart, periodEnd time.Time) (map[uuid.UUID]float64, error) {
	result := make(map[uuid.UUID]float64)
	if len(categoryIDs) == 0 {
		return result, nil
	}

	type catSpend struct {
		CategoryID uuid.UUID `db:"category_id"`
		Total      float64   `db:"total"`
	}

	query, args, err := sqlx.In(`SELECT category_id, COALESCE(SUM(CAST(amount AS NUMERIC)), 0) as total
		FROM transactions
		WHERE user_id = ? AND category_id IN (?) AND type = 'expense'
		AND date >= ? AND date < ? AND deleted_at IS NULL
		GROUP BY category_id`, userID, categoryIDs, periodStart, periodEnd)
	if err != nil {
		return result, err
	}
	query = r.db.Rebind(query)

	var rows []catSpend
	if err := r.db.Select(&rows, query, args...); err != nil {
		return result, err
	}

	for _, r := range rows {
		result[r.CategoryID] = r.Total
	}
	return result, nil
}
