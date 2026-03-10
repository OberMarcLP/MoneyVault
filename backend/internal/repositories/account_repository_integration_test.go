package repositories_test

import (
	"testing"

	"moneyvault/internal/repositories"
	"moneyvault/tests/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_AccountRepository(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	userRepo := repositories.NewUserRepository(db)
	repo := repositories.NewAccountRepository(db)

	t.Run("Create and GetByID", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, userRepo, "acct@test.com")
		acct := testhelpers.CreateTestAccount(t, repo, user.ID, "Checking")

		got, err := repo.GetByID(acct.ID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Checking", got.Name)
		assert.Equal(t, "1000.00", got.Balance)
		assert.True(t, got.IsActive)
	})

	t.Run("ListByUser returns only that user's accounts", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user1 := testhelpers.CreateTestUser(t, userRepo, "user1@test.com")
		user2 := testhelpers.CreateTestUser(t, userRepo, "user2@test.com")
		testhelpers.CreateTestAccount(t, repo, user1.ID, "User1 Account")
		testhelpers.CreateTestAccount(t, repo, user1.ID, "User1 Savings")
		testhelpers.CreateTestAccount(t, repo, user2.ID, "User2 Account")

		accts, err := repo.ListByUser(user1.ID)
		require.NoError(t, err)
		assert.Len(t, accts, 2)

		accts2, err := repo.ListByUser(user2.ID)
		require.NoError(t, err)
		assert.Len(t, accts2, 1)
	})

	t.Run("GetByID with wrong userID fails", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, userRepo, "owner@test.com")
		acct := testhelpers.CreateTestAccount(t, repo, user.ID, "Private")

		_, err := repo.GetByID(acct.ID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("Update name and balance", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, userRepo, "update@test.com")
		acct := testhelpers.CreateTestAccount(t, repo, user.ID, "Old Name")

		acct.Name = "New Name"
		acct.Balance = "2500.50"
		err := repo.Update(acct)
		require.NoError(t, err)

		got, err := repo.GetByID(acct.ID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", got.Name)
		assert.Equal(t, "2500.50", got.Balance)
	})

	t.Run("Delete and verify gone", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, userRepo, "delacc@test.com")
		acct := testhelpers.CreateTestAccount(t, repo, user.ID, "Temporary")

		err := repo.Delete(acct.ID, user.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(acct.ID, user.ID)
		assert.Error(t, err)
	})

	t.Run("Cascade: deleting user removes accounts", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, userRepo, "cascade@test.com")
		testhelpers.CreateTestAccount(t, repo, user.ID, "Will Be Gone")

		err := userRepo.Delete(user.ID)
		require.NoError(t, err)

		accts, err := repo.ListByUser(user.ID)
		require.NoError(t, err)
		assert.Len(t, accts, 0)
	})
}
