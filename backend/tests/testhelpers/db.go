package testhelpers

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDSN returns the test database connection string.
// Override with TEST_DATABASE_URL env var if needed.
func TestDSN() string {
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgres://moneyvault_test:moneyvault_test@localhost:5433/moneyvault_test?sslmode=disable"
}

// MigrationsPath returns the file:// path to the migrations directory,
// resolved relative to this source file so it works from any working directory.
func MigrationsPath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	migrationsDir := filepath.Join(dir, "..", "..", "migrations")
	migrationsDir = filepath.ToSlash(filepath.Clean(migrationsDir))
	return "file://" + migrationsDir
}

// SetupTestDB connects to the test database and runs all migrations.
// Skips the test when -short flag is set so unit tests work without a DB.
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test (requires test database)")
	}

	dsn := TestDSN()
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run migrations
	m, err := migrate.New(MigrationsPath(), dsn)
	if err != nil {
		db.Close()
		t.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TruncateTables truncates all application tables in a single atomic statement
// to avoid deadlocks when multiple test packages run in parallel.
func TruncateTables(t *testing.T, db *sqlx.DB) {
	t.Helper()
	tables := []string{
		"push_subscriptions",
		"exchange_connections",
		"dividends",
		"webauthn_credentials",
		"dek_sessions",
		"audit_logs",
		"password_reset_tokens",
		"revoked_tokens",
		"alert_rules",
		"notifications",
		"net_worth_snapshots",
		"wallet_transactions",
		"wallets",
		"trade_lots",
		"price_history",
		"price_cache",
		"holdings",
		"import_jobs",
		"recurring_transactions",
		"budgets",
		"transactions",
		"categories",
		"accounts",
		"users",
	}
	query := "TRUNCATE TABLE "
	for i, table := range tables {
		if i > 0 {
			query += ", "
		}
		query += table
	}
	query += " CASCADE"
	_, err := db.Exec(query)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}
