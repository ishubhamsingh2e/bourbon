// Package contracts provides reusable, standardized compliance test suites for Bourbon pluggable adapters.
package contracts

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/database"
)

// VerifyDatabaseContract executes a standard compliance test suite against any database.DB implementation.
func VerifyDatabaseContract(t *testing.T, db database.DB) {
	ctx := context.Background()

	t.Run("Ping", func(t *testing.T) {
		if err := db.Ping(ctx); err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
	})

	t.Run("Exec", func(t *testing.T) {
		res, err := db.Exec(ctx, "contract_test_exec", "sample_val")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil sql.Result")
		}
	})

	t.Run("Transaction_Commit", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}
		_, err = tx.Exec(ctx, "contract_tx_commit", "committed")
		if err != nil {
			t.Fatalf("tx.Exec failed: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("tx.Commit failed: %v", err)
		}
	})

	t.Run("Transaction_Rollback", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("BeginTx failed: %v", err)
		}
		_, err = tx.Exec(ctx, "contract_tx_rollback", "staged")
		if err != nil {
			t.Fatalf("tx.Exec failed: %v", err)
		}
		if err := tx.Rollback(); err != nil {
			t.Fatalf("tx.Rollback failed: %v", err)
		}
	})

	t.Run("WithTx_Helper", func(t *testing.T) {
		err := database.WithTx(ctx, db, func(ctx context.Context, tx database.Tx) error {
			_, err := tx.Exec(ctx, "contract_with_tx", "val")
			return err
		})
		if err != nil {
			t.Fatalf("WithTx failed: %v", err)
		}
	})

	t.Run("Replica", func(t *testing.T) {
		replica := db.Replica()
		if replica == nil {
			t.Fatalf("Replica() returned nil")
		}
	})
}
