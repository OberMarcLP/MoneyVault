package repositories

import (
	"time"

	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AnalyticsRepository struct {
	db *sqlx.DB
}

func NewAnalyticsRepository(db *sqlx.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// Net worth snapshots
func (r *AnalyticsRepository) UpsertNetWorthSnapshot(s *models.NetWorthSnapshot) error {
	query := `INSERT INTO net_worth_snapshots (id, user_id, date, total_value, accounts_value, investments_value, crypto_value, breakdown)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, date) DO UPDATE SET
		total_value = $4, accounts_value = $5, investments_value = $6, crypto_value = $7, breakdown = $8`
	_, err := r.db.Exec(query, s.ID, s.UserID, s.Date, s.TotalValue, s.AccountsValue, s.InvestmentsValue, s.CryptoValue, s.Breakdown)
	return err
}

func (r *AnalyticsRepository) GetNetWorthHistory(userID uuid.UUID, days int) ([]models.NetWorthSnapshot, error) {
	var snapshots []models.NetWorthSnapshot
	err := r.db.Select(&snapshots,
		`SELECT * FROM net_worth_snapshots WHERE user_id = $1 AND date >= NOW() - INTERVAL '1 day' * $2 ORDER BY date ASC`,
		userID, days)
	if snapshots == nil {
		snapshots = []models.NetWorthSnapshot{}
	}
	return snapshots, err
}

func (r *AnalyticsRepository) GetLatestNetWorth(userID uuid.UUID) (*models.NetWorthSnapshot, error) {
	var s models.NetWorthSnapshot
	err := r.db.Get(&s, `SELECT * FROM net_worth_snapshots WHERE user_id = $1 ORDER BY date DESC LIMIT 1`, userID)
	return &s, err
}

// Spending analytics
type spendingRow struct {
	CategoryID uuid.UUID `db:"category_id"`
	Total      float64   `db:"total"`
	Count      int       `db:"count"`
}

func (r *AnalyticsRepository) GetSpendingByCategory(userID uuid.UUID, from, to time.Time) ([]spendingRow, error) {
	var rows []spendingRow
	err := r.db.Select(&rows,
		`SELECT category_id, COALESCE(SUM(CAST(amount AS NUMERIC)), 0) as total, COUNT(*) as count
		FROM transactions
		WHERE user_id = $1 AND type = 'expense' AND date >= $2 AND date < $3 AND category_id IS NOT NULL
		GROUP BY category_id ORDER BY total DESC`,
		userID, from, to)
	if rows == nil {
		rows = []spendingRow{}
	}
	return rows, err
}

type trendRow struct {
	Period  string  `db:"period"`
	Type    string  `db:"type"`
	Total   float64 `db:"total"`
}

func (r *AnalyticsRepository) GetMonthlyTrends(userID uuid.UUID, months int) ([]trendRow, error) {
	var rows []trendRow
	err := r.db.Select(&rows,
		`SELECT TO_CHAR(date, 'YYYY-MM') as period, type, COALESCE(SUM(CAST(amount AS NUMERIC)), 0) as total
		FROM transactions
		WHERE user_id = $1 AND date >= (NOW() - INTERVAL '1 month' * $2)::date AND type IN ('income', 'expense')
		GROUP BY TO_CHAR(date, 'YYYY-MM'), type
		ORDER BY period ASC`,
		userID, months)
	if rows == nil {
		rows = []trendRow{}
	}
	return rows, err
}

type topExpenseRow struct {
	Description string    `db:"description"`
	Amount      float64   `db:"amount"`
	Date        time.Time `db:"date"`
	CategoryID  *uuid.UUID `db:"category_id"`
}

func (r *AnalyticsRepository) GetTopExpenses(userID uuid.UUID, from, to time.Time, limit int) ([]topExpenseRow, error) {
	var rows []topExpenseRow
	err := r.db.Select(&rows,
		`SELECT description, CAST(amount AS NUMERIC) as amount, date, category_id
		FROM transactions
		WHERE user_id = $1 AND type = 'expense' AND date >= $2 AND date < $3
		ORDER BY CAST(amount AS NUMERIC) DESC
		LIMIT $4`,
		userID, from, to, limit)
	if rows == nil {
		rows = []topExpenseRow{}
	}
	return rows, err
}

type AccountBalanceRow struct {
	Type     string `db:"type"`
	Balance  string `db:"balance"`
	Currency string `db:"currency"`
}

// Account balances for snapshot (returns raw/encrypted balances for decryption in service layer)
func (r *AnalyticsRepository) GetAccountBalances(userID uuid.UUID) ([]AccountBalanceRow, error) {
	var rows []AccountBalanceRow
	err := r.db.Select(&rows,
		`SELECT type, balance, currency FROM accounts WHERE user_id = $1 AND is_active = true`,
		userID)
	if rows == nil {
		rows = []AccountBalanceRow{}
	}
	return rows, err
}

// Get all distinct user IDs for batch snapshot
func (r *AnalyticsRepository) GetAllUserIDs() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Select(&ids, `SELECT id FROM users`)
	if ids == nil {
		ids = []uuid.UUID{}
	}
	return ids, err
}

// Monthly income/expense for cash flow
func (r *AnalyticsRepository) GetMonthlyAverages(userID uuid.UUID, months int) (income float64, expenses float64, err error) {
	type result struct {
		Type  string  `db:"type"`
		Avg   float64 `db:"avg"`
	}
	var rows []result
	err = r.db.Select(&rows,
		`SELECT type, COALESCE(AVG(monthly_total), 0) as avg FROM (
			SELECT type, TO_CHAR(date, 'YYYY-MM') as month, SUM(CAST(amount AS NUMERIC)) as monthly_total
			FROM transactions
			WHERE user_id = $1 AND type IN ('income', 'expense') AND date >= (NOW() - INTERVAL '1 month' * $2)::date
			GROUP BY type, TO_CHAR(date, 'YYYY-MM')
		) sub GROUP BY type`,
		userID, months)
	for _, r := range rows {
		if r.Type == "income" {
			income = r.Avg
		} else if r.Type == "expense" {
			expenses = r.Avg
		}
	}
	return
}
