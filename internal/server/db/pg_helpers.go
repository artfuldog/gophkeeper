package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/jackc/pgx/v4"
)

// runBatch is helper function to run sql requests in batches.
//
// runBatch does it job in one transaction and checks results of every request.
func (db *DBPosgtre) runBatch(ctx context.Context, batch *pgx.Batch, componentName string) error {
	tx, err := db.beginTx(ctx, componentName)
	if err != nil {
		return err
	}
	defer db.deferTxRollback(ctx, tx)

	batchRes := tx.SendBatch(ctx, batch)
	defer batchRes.Close()

	for i := 0; i < batch.Len(); i++ {
		ct, err := batchRes.Exec()
		if wrappedErr := wrapPgError(err); wrappedErr != nil {
			return wrappedErr
		}
		if ct.RowsAffected() < 1 {
			return stackErrors(ErrOperationFailed, errors.New("no rows affected"))
		}
	}

	batchRes.Close()

	if err := db.commitTx(ctx, tx, componentName); err != nil {
		return err
	}

	return nil
}

// stackErrors is helper function to wrap database error.
//
// Wraps internal with database error.
func stackErrors(dbErr error, intErr error) error {
	return fmt.Errorf("%w::%v", dbErr, intErr)
}

// getHashUpdatedItem is a helper function for generate hash and updated field
// value for new or updated item
func getHashUpdatedItem(itemName, itemType string) (updated string, hash []byte) {
	updated = time.Now().Format(time.RFC3339)
	hash = crypt.GetMD5hash(fmt.Sprintf(("%s:%s:%v"), itemName, itemType, updated))
	return
}

// beginTx is a helper function to start transaction.
//
// In case of failure logs an error, returns empty pgx.Tx and already stacked error.
// componentName is used in log's record.
func (db *DBPosgtre) beginTx(ctx context.Context, componentName string) (tx pgx.Tx, err error) {
	if tx, err = db.pool.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		db.logger.Error(err, "begin transaction", componentName)
		return nil, stackErrors(ErrTransactionFailed, err)
	}
	return
}

// beginTxRO is a helper function to start Read-Only transaction.
//
// In case of failure logs an error, returns empty pgx.Tx and already stacked error.
// componentName is used in log's record.
func (db *DBPosgtre) beginTxRO(ctx context.Context, componentName string) (tx pgx.Tx, err error) {
	if tx, err = db.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly}); err != nil {
		db.logger.Error(err, "begin transaction", componentName)
		return nil, stackErrors(ErrTransactionFailed, err)
	}
	return
}

// commitTx is a helper function for commit transaction
//
// In case of failure logs an error and returns ErrTransactionFailed.
// componentName is used in log's record.
func (db *DBPosgtre) commitTx(ctx context.Context, tx pgx.Tx, componentName string) (err error) {
	if err = tx.Commit(ctx); err != nil {
		db.logger.Error(err, "commit transaction", componentName)
		return stackErrors(ErrTransactionFailed, err)
	}
	return
}

// deferTxRollback is a helper function for defering transaction rollback
//
// In case of failure logs an error.
// componentName is used in log's record.
func (db *DBPosgtre) deferTxRollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
		db.logger.Error(err, "failed", "deferTxRollback")
	}
}
