package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(account *models.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, name, type, currency, balance, institution, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at`
	return r.db.QueryRow(query,
		account.ID, account.UserID, account.Name, account.Type,
		account.Currency, account.Balance, account.Institution, account.IsActive,
	).Scan(&account.CreatedAt, &account.UpdatedAt)
}

func (r *AccountRepository) GetByID(id, userID uuid.UUID) (*models.Account, error) {
	var account models.Account
	err := r.db.Get(&account, "SELECT * FROM accounts WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL", id, userID)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) ListByUser(userID uuid.UUID) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Select(&accounts, "SELECT * FROM accounts WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC", userID)
	return accounts, err
}

func (r *AccountRepository) Update(account *models.Account) error {
	query := `
		UPDATE accounts SET name = $3, type = $4, currency = $5,
		balance = $6, institution = $7, is_active = $8, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(query,
		account.ID, account.UserID, account.Name, account.Type,
		account.Currency, account.Balance, account.Institution, account.IsActive,
	)
	return err
}

func (r *AccountRepository) Delete(id, userID uuid.UUID) error {
	_, err := r.db.Exec("UPDATE accounts SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL", id, userID)
	return err
}
