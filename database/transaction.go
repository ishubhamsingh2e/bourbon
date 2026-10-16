package database

import (
	"context"
	"database/sql"
	"fmt"
)

// TxFunc defines the callback executed within a managed transaction.
type TxFunc func(ctx context.Context, tx Tx) error

// WithTx executes fn within a managed transaction, committing on success or rolling back on error or panic.
func WithTx(ctx context.Context, db DB, fn TxFunc, opts ...*sql.TxOptions) (err error) {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	tx, err := db.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("database: failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("database: failed to commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(ctx, tx)
	return err
}
