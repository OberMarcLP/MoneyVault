package repositories_test

import (
	"testing"
	"time"

	"moneyvault/internal/repositories"
	"moneyvault/tests/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_UserRepository(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Create and GetByID round-trip", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, repo, "roundtrip@test.com")

		got, err := repo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, got.ID)
		assert.Equal(t, "roundtrip@test.com", got.Email)
		assert.False(t, got.CreatedAt.IsZero())
	})

	t.Run("GetByEmail found and not-found", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		testhelpers.CreateTestUser(t, repo, "findme@test.com")

		got, err := repo.GetByEmail("findme@test.com")
		require.NoError(t, err)
		assert.Equal(t, "findme@test.com", got.Email)

		_, err = repo.GetByEmail("nonexistent@test.com")
		assert.Error(t, err)
	})

	t.Run("Update email", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, repo, "before@test.com")
		user.Email = "after@test.com"
		err := repo.Update(user)
		require.NoError(t, err)

		got, err := repo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "after@test.com", got.Email)
	})

	t.Run("Delete and verify gone", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, repo, "deleteme@test.com")
		err := repo.Delete(user.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(user.ID)
		assert.Error(t, err)
	})

	t.Run("Count", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		count, err := repo.Count()
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		testhelpers.CreateTestUser(t, repo, "one@test.com")
		testhelpers.CreateTestUser(t, repo, "two@test.com")

		count, err = repo.Count()
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("IncrementFailedAttempts", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, repo, "lockout@test.com")

		attempts, err := repo.IncrementFailedAttempts(user.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, attempts)

		attempts, err = repo.IncrementFailedAttempts(user.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("ResetFailedAttempts", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, repo, "reset@test.com")
		_, _ = repo.IncrementFailedAttempts(user.ID)
		_, _ = repo.IncrementFailedAttempts(user.ID)
		lockTime := time.Now().Add(30 * time.Minute)
		_ = repo.LockUser(user.ID, lockTime)

		err := repo.ResetFailedAttempts(user.ID)
		require.NoError(t, err)

		got, err := repo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, got.FailedLoginAttempts)
		assert.Nil(t, got.LockedUntil)
	})

	t.Run("LockUser", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		user := testhelpers.CreateTestUser(t, repo, "lockuser@test.com")
		lockTime := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)
		err := repo.LockUser(user.ID, lockTime)
		require.NoError(t, err)

		got, err := repo.GetByID(user.ID)
		require.NoError(t, err)
		require.NotNil(t, got.LockedUntil)
		assert.WithinDuration(t, lockTime, *got.LockedUntil, time.Second)
	})

	t.Run("List returns all users", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		testhelpers.CreateTestUser(t, repo, "a@test.com")
		testhelpers.CreateTestUser(t, repo, "b@test.com")

		users, err := repo.List()
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("Duplicate email fails", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		testhelpers.CreateTestUser(t, repo, "unique@test.com")
		// Attempt to create another user with the same email
		dupe := testhelpers.CreateDummyUser("unique@test.com")
		err := repo.Create(dupe)
		assert.Error(t, err)
	})

	t.Run("GetByID with non-existent ID returns error", func(t *testing.T) {
		testhelpers.TruncateTables(t, db)

		_, err := repo.GetByID(uuid.New())
		assert.Error(t, err)
	})
}
