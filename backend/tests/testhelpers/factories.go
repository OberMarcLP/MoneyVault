package testhelpers

import (
	"encoding/json"
	"testing"
	"time"

	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	"github.com/google/uuid"
)

// CreateDummyUser returns a User struct (not yet inserted) for testing scenarios
// like duplicate email validation.
func CreateDummyUser(email string) *models.User {
	return &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$dummysalt$dummyhash",
		Role:         models.RoleUser,
		EncryptedDEK: "test-encrypted-dek",
		KEKSalt:      "test-kek-salt",
		Preferences:  json.RawMessage(`{"theme":"system","currency":"USD","locale":"en-US"}`),
	}
}

// CreateTestUser inserts a user with a dummy password hash and DEK.
func CreateTestUser(t *testing.T, repo *repositories.UserRepository, email string) *models.User {
	t.Helper()
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$dummysalt$dummyhash",
		Role:         models.RoleUser,
		EncryptedDEK: "test-encrypted-dek",
		KEKSalt:      "test-kek-salt",
		Preferences:  json.RawMessage(`{"theme":"system","currency":"USD","locale":"en-US"}`),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

// CreateTestAccount inserts a checking account with the given name and $1000 balance.
func CreateTestAccount(t *testing.T, repo *repositories.AccountRepository, userID uuid.UUID, name string) *models.Account {
	t.Helper()
	account := &models.Account{
		ID:       uuid.New(),
		UserID:   userID,
		Name:     name,
		Type:     models.AccountChecking,
		Currency: "USD",
		Balance:  "1000.00",
		IsActive: true,
	}
	if err := repo.Create(account); err != nil {
		t.Fatalf("failed to create test account: %v", err)
	}
	return account
}

// CreateTestTransaction inserts a transaction.
func CreateTestTransaction(
	t *testing.T,
	repo *repositories.TransactionRepository,
	userID, accountID uuid.UUID,
	txType models.TransactionType,
	amount, description string,
	date time.Time,
) *models.Transaction {
	t.Helper()
	tx := &models.Transaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		UserID:      userID,
		Type:        txType,
		Amount:      amount,
		Currency:    "USD",
		Description: description,
		Date:        date,
		Tags:        json.RawMessage(`[]`),
	}
	if err := repo.Create(tx); err != nil {
		t.Fatalf("failed to create test transaction: %v", err)
	}
	return tx
}
