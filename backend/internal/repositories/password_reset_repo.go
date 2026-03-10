package repositories

import (
	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PasswordResetRepository struct {
	db *sqlx.DB
}

func NewPasswordResetRepository(db *sqlx.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(token *models.PasswordResetToken) error {
	query := `INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := r.db.Exec(query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.Used)
	return err
}

func (r *PasswordResetRepository) FindByTokenHash(hash string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	err := r.db.Get(&token, `SELECT * FROM password_reset_tokens WHERE token_hash = $1`, hash)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *PasswordResetRepository) MarkUsed(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE password_reset_tokens SET used = TRUE WHERE id = $1`, id)
	return err
}

func (r *PasswordResetRepository) CleanupExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used = TRUE`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
