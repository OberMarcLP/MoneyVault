package repositories_test

import (
	"testing"
	"time"

	"moneyvault/internal/models"
	"moneyvault/internal/repositories"
	"moneyvault/tests/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_TransactionRepository(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	userRepo := repositories.NewUserRepository(db)
	acctRepo := repositories.NewAccountRepository(db)
	repo := repositories.NewTransactionRepository(db)

	// Helper to set up a user + account for each subtest
	setup := func(t *testing.T) (uuid.UUID, uuid.UUID) {
		t.Helper()
		testhelpers.TruncateTables(t, db)
		user := testhelpers.CreateTestUser(t, userRepo, "tx@test.com")
		acct := testhelpers.CreateTestAccount(t, acctRepo, user.ID, "Main")
		return user.ID, acct.ID
	}

	t.Run("Create and GetByID", func(t *testing.T) {
		userID, acctID := setup(t)

		tx := testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "50.00", "Coffee", time.Now())

		got, err := repo.GetByID(tx.ID, userID)
		require.NoError(t, err)
		assert.Equal(t, "50.00", got.Amount)
		assert.Equal(t, "Coffee", got.Description)
		assert.Equal(t, models.TransactionExpense, got.Type)
	})

	t.Run("GetByID with wrong userID fails", func(t *testing.T) {
		userID, acctID := setup(t)

		tx := testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "25.00", "Snack", time.Now())

		_, err := repo.GetByID(tx.ID, uuid.New())
		assert.Error(t, err)
	})

	t.Run("List with no filters", func(t *testing.T) {
		userID, acctID := setup(t)

		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "10.00", "One", time.Now())
		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionIncome, "20.00", "Two", time.Now())

		filter := models.TransactionFilter{Page: 1, PerPage: 50}
		txs, total, err := repo.List(userID, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, txs, 2)
	})

	t.Run("List filtered by Type", func(t *testing.T) {
		userID, acctID := setup(t)

		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "30.00", "Expense1", time.Now())
		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionIncome, "40.00", "Income1", time.Now())
		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "50.00", "Expense2", time.Now())

		expenseType := models.TransactionExpense
		filter := models.TransactionFilter{Type: &expenseType, Page: 1, PerPage: 50}
		txs, total, err := repo.List(userID, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, txs, 2)
		for _, tx := range txs {
			assert.Equal(t, models.TransactionExpense, tx.Type)
		}
	})

	t.Run("List filtered by AccountID", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, userRepo, "multiacct@test.com")
		acct1 := testhelpers.CreateTestAccount(t, acctRepo, user.ID, "Account1")
		acct2 := testhelpers.CreateTestAccount(t, acctRepo, user.ID, "Account2")

		testhelpers.CreateTestTransaction(t, repo, user.ID, acct1.ID,
			models.TransactionExpense, "10.00", "From Acct1", time.Now())
		testhelpers.CreateTestTransaction(t, repo, user.ID, acct2.ID,
			models.TransactionExpense, "20.00", "From Acct2", time.Now())

		filter := models.TransactionFilter{AccountID: &acct1.ID, Page: 1, PerPage: 50}
		txs, total, err := repo.List(user.ID, filter)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, txs, 1)
		assert.Equal(t, acct1.ID, txs[0].AccountID)
	})

	t.Run("List filtered by DateFrom and DateTo", func(t *testing.T) {
		userID, acctID := setup(t)

		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "10.00", "Old", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "20.00", "Mid", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
		testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "30.00", "Recent", time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC))

		from := "2025-06-01"
		to := "2025-12-31"
		filter := models.TransactionFilter{DateFrom: &from, DateTo: &to, Page: 1, PerPage: 50}
		txs, total, err := repo.List(userID, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, txs, 2)
	})

	t.Run("Update amount and description", func(t *testing.T) {
		userID, acctID := setup(t)

		tx := testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "10.00", "Before", time.Now())

		tx.Amount = "99.99"
		tx.Description = "After"
		err := repo.Update(tx)
		require.NoError(t, err)

		got, err := repo.GetByID(tx.ID, userID)
		require.NoError(t, err)
		assert.Equal(t, "99.99", got.Amount)
		assert.Equal(t, "After", got.Description)
	})

	t.Run("Delete and verify gone", func(t *testing.T) {
		userID, acctID := setup(t)

		tx := testhelpers.CreateTestTransaction(t, repo, userID, acctID,
			models.TransactionExpense, "5.00", "Bye", time.Now())

		err := repo.Delete(tx.ID, userID)
		require.NoError(t, err)

		_, err = repo.GetByID(tx.ID, userID)
		assert.Error(t, err)
	})
}
