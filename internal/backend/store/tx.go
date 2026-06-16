package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// WithTx runs fn inside a transaction.
func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	return db.WithTxOptions(ctx, nil, fn)
}

// WithTxOptions runs fn inside a transaction with explicit options.
func (db *DB) WithTxOptions(ctx context.Context, opts *sql.TxOptions, fn func(*sql.Tx) error) (err error) {
	if db == nil || db.SQL == nil {
		return errors.New("store: nil database")
	}
	if fn == nil {
		return errors.New("store: transaction function is required")
	}

	tx, err := db.SQL.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}

		if err != nil {
			_ = tx.Rollback()
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("commit transaction: %w", commitErr)
		}
	}()

	err = fn(tx)
	return err
}
