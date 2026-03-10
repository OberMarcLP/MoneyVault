package repositories

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RunInTransaction executes fn within a database transaction.
// If fn returns an error, the transaction is rolled back. Otherwise, it is committed.
func RunInTransaction(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
