package services

import (
	"encoding/base64"
	"errors"
	"fmt"

	"moneyvault/internal/encryption"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/argon2"
)

type E2EService struct {
	db       *sqlx.DB
	userRepo *repositories.UserRepository
	enc      *encryption.Service
}

func NewE2EService(db *sqlx.DB, userRepo *repositories.UserRepository, enc *encryption.Service) *E2EService {
	return &E2EService{db: db, userRepo: userRepo, enc: enc}
}

// ExportData returns all user data with server-side decrypted sensitive fields.
func (s *E2EService) ExportData(userID uuid.UUID) (*models.E2EMigrateDataRequest, error) {
	var accounts []struct {
		ID      uuid.UUID `db:"id"`
		Name    string    `db:"name"`
		Balance string    `db:"balance"`
	}
	if err := s.db.Select(&accounts, "SELECT id, name, balance FROM accounts WHERE user_id = $1", userID); err != nil {
		return nil, err
	}

	var transactions []struct {
		ID          uuid.UUID `db:"id"`
		Amount      string    `db:"amount"`
		Description string    `db:"description"`
	}
	if err := s.db.Select(&transactions, "SELECT id, amount, description FROM transactions WHERE user_id = $1", userID); err != nil {
		return nil, err
	}

	result := &models.E2EMigrateDataRequest{}

	for _, a := range accounts {
		name, _ := s.enc.DecryptField(userID, a.Name)
		balance, _ := s.enc.DecryptField(userID, a.Balance)
		result.Accounts = append(result.Accounts, models.E2EMigrateAccount{
			ID:      a.ID.String(),
			Name:    name,
			Balance: balance,
		})
	}

	for _, t := range transactions {
		amount, _ := s.enc.DecryptField(userID, t.Amount)
		desc, _ := s.enc.DecryptField(userID, t.Description)
		result.Transactions = append(result.Transactions, models.E2EMigrateTransaction{
			ID:          t.ID.String(),
			Amount:      amount,
			Description: desc,
		})
	}

	return result, nil
}

// MigrateAndEnable atomically updates all user data with client-encrypted values and enables E2E.
func (s *E2EService) MigrateAndEnable(userID uuid.UUID, password, e2eEncryptedDEK, e2eKEKSalt string, data models.E2EMigrateDataRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify password
	salt, err := base64.StdEncoding.DecodeString(user.KEKSalt)
	if err != nil {
		return errors.New("internal error")
	}
	passwordHash := hashPassword(password, salt)
	storedHash, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		return errors.New("internal error")
	}
	if !compareHashBytes(passwordHash, storedHash) {
		return errors.New("invalid password")
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update all accounts with client-encrypted data
	for _, a := range data.Accounts {
		if _, err := tx.Exec("UPDATE accounts SET name = $1, balance = $2 WHERE id = $3 AND user_id = $4",
			a.Name, a.Balance, a.ID, userID); err != nil {
			return fmt.Errorf("failed to update account: %w", err)
		}
	}

	// Update all transactions with client-encrypted data
	for _, t := range data.Transactions {
		if _, err := tx.Exec("UPDATE transactions SET amount = $1, description = $2 WHERE id = $3 AND user_id = $4",
			t.Amount, t.Description, t.ID, userID); err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	// Enable E2E
	if _, err := tx.Exec("UPDATE users SET e2e_enabled = TRUE, e2e_encrypted_dek = $1, e2e_kek_salt = $2, updated_at = NOW() WHERE id = $3",
		e2eEncryptedDEK, e2eKEKSalt, userID); err != nil {
		return fmt.Errorf("failed to enable E2E: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	s.enc.SetE2EUser(userID, true)
	return nil
}

// MigrateAndDisable re-encrypts all user data with server-side DEK and disables E2E.
func (s *E2EService) MigrateAndDisable(userID uuid.UUID, data models.E2EMigrateDataRequest) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Temporarily disable E2E so EncryptField works normally
	s.enc.SetE2EUser(userID, false)

	// Re-encrypt all accounts with server-side DEK
	for _, a := range data.Accounts {
		name, err := s.enc.EncryptField(userID, a.Name)
		if err != nil {
			return fmt.Errorf("failed to encrypt account name: %w", err)
		}
		balance, err := s.enc.EncryptField(userID, a.Balance)
		if err != nil {
			return fmt.Errorf("failed to encrypt account balance: %w", err)
		}
		if _, err := tx.Exec("UPDATE accounts SET name = $1, balance = $2 WHERE id = $3 AND user_id = $4",
			name, balance, a.ID, userID); err != nil {
			return fmt.Errorf("failed to update account: %w", err)
		}
	}

	// Re-encrypt all transactions with server-side DEK
	for _, t := range data.Transactions {
		amount, err := s.enc.EncryptField(userID, t.Amount)
		if err != nil {
			return fmt.Errorf("failed to encrypt amount: %w", err)
		}
		desc, err := s.enc.EncryptField(userID, t.Description)
		if err != nil {
			return fmt.Errorf("failed to encrypt description: %w", err)
		}
		if _, err := tx.Exec("UPDATE transactions SET amount = $1, description = $2 WHERE id = $3 AND user_id = $4",
			amount, desc, t.ID, userID); err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	// Disable E2E
	if _, err := tx.Exec("UPDATE users SET e2e_enabled = FALSE, e2e_encrypted_dek = '', e2e_kek_salt = '', updated_at = NOW() WHERE id = $1",
		userID); err != nil {
		s.enc.SetE2EUser(userID, true) // Restore on failure
		return fmt.Errorf("failed to disable E2E: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.enc.SetE2EUser(userID, true) // Restore on failure
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func hashPassword(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
}

func compareHashBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
