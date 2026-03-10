package repositories

import (
	"time"

	"moneyvault/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, totp_enabled, encrypted_dek, kek_salt, preferences, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at`
	return r.db.QueryRow(query,
		user.ID, user.Email, user.PasswordHash, user.Role,
		user.TOTPEnabled, user.EncryptedDEK, user.KEKSalt, user.Preferences,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET email = $2, password_hash = $3, role = $4,
		totp_secret = $5, totp_enabled = $6, encrypted_dek = $7,
		kek_salt = $8, preferences = $9, email_verified = $10, updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.Exec(query,
		user.ID, user.Email, user.PasswordHash, user.Role,
		user.TOTPSecret, user.TOTPEnabled, user.EncryptedDEK,
		user.KEKSalt, user.Preferences, user.EmailVerified,
	)
	return err
}

func (r *UserRepository) List() ([]models.User, error) {
	var users []models.User
	err := r.db.Select(&users, "SELECT * FROM users ORDER BY created_at DESC")
	return users, err
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (r *UserRepository) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM users")
	return count, err
}

func (r *UserRepository) IncrementFailedAttempts(userID uuid.UUID) (int, error) {
	var attempts int
	err := r.db.Get(&attempts,
		`UPDATE users SET failed_login_attempts = failed_login_attempts + 1, updated_at = NOW()
		 WHERE id = $1 RETURNING failed_login_attempts`, userID)
	return attempts, err
}

func (r *UserRepository) ResetFailedAttempts(userID uuid.UUID) error {
	_, err := r.db.Exec(
		`UPDATE users SET failed_login_attempts = 0, locked_until = NULL, updated_at = NOW() WHERE id = $1`, userID)
	return err
}

func (r *UserRepository) LockUser(userID uuid.UUID, until time.Time) error {
	_, err := r.db.Exec(
		`UPDATE users SET locked_until = $2, updated_at = NOW() WHERE id = $1`, userID, until)
	return err
}

func (r *UserRepository) UpdateE2E(userID uuid.UUID, enabled bool, encryptedDEK, kekSalt string) error {
	_, err := r.db.Exec(
		`UPDATE users SET e2e_enabled = $2, e2e_encrypted_dek = $3, e2e_kek_salt = $4, updated_at = NOW() WHERE id = $1`,
		userID, enabled, encryptedDEK, kekSalt)
	return err
}
