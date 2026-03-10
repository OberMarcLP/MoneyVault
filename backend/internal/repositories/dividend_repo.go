package repositories

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"moneyvault/internal/models"
)

type DividendRepository struct {
	db *sqlx.DB
}

func NewDividendRepository(db *sqlx.DB) *DividendRepository {
	return &DividendRepository{db: db}
}

func (r *DividendRepository) Create(d *models.Dividend) error {
	_, err := r.db.NamedExec(`
		INSERT INTO dividends (id, holding_id, user_id, amount, currency, ex_date, pay_date, notes)
		VALUES (:id, :holding_id, :user_id, :amount, :currency, :ex_date, :pay_date, :notes)
	`, d)
	return err
}

func (r *DividendRepository) GetByID(id, userID uuid.UUID) (*models.Dividend, error) {
	var d models.Dividend
	err := r.db.Get(&d, `SELECT * FROM dividends WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return &d, err
}

func (r *DividendRepository) ListByUser(userID uuid.UUID) ([]models.Dividend, error) {
	var divs []models.Dividend
	err := r.db.Select(&divs, `SELECT * FROM dividends WHERE user_id = $1 AND deleted_at IS NULL ORDER BY ex_date DESC`, userID)
	if divs == nil {
		divs = []models.Dividend{}
	}
	return divs, err
}

func (r *DividendRepository) ListByHolding(holdingID, userID uuid.UUID) ([]models.Dividend, error) {
	var divs []models.Dividend
	err := r.db.Select(&divs, `SELECT * FROM dividends WHERE holding_id = $1 AND user_id = $2 AND deleted_at IS NULL ORDER BY ex_date DESC`, holdingID, userID)
	if divs == nil {
		divs = []models.Dividend{}
	}
	return divs, err
}

func (r *DividendRepository) Update(d *models.Dividend) error {
	_, err := r.db.NamedExec(`
		UPDATE dividends SET amount = :amount, currency = :currency, ex_date = :ex_date, pay_date = :pay_date, notes = :notes
		WHERE id = :id AND user_id = :user_id AND deleted_at IS NULL
	`, d)
	return err
}

func (r *DividendRepository) Delete(id, userID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE dividends SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	return err
}

func (r *DividendRepository) GetTotalByUser(userID uuid.UUID) (float64, error) {
	var total *float64
	err := r.db.Get(&total, `SELECT SUM(CAST(amount AS NUMERIC)) FROM dividends WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	if total == nil {
		return 0, err
	}
	return *total, err
}

func (r *DividendRepository) GetTotalYTD(userID uuid.UUID, year int) (float64, error) {
	var total *float64
	err := r.db.Get(&total, `
		SELECT SUM(CAST(amount AS NUMERIC)) FROM dividends
		WHERE user_id = $1 AND EXTRACT(YEAR FROM ex_date) = $2 AND deleted_at IS NULL
	`, userID, year)
	if total == nil {
		return 0, err
	}
	return *total, err
}
