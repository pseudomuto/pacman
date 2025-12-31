package data

import (
	"context"
	"fmt"

	"github.com/pseudomuto/pacman/internal/ent"
)

// WithTx starts a new ent transaction, invokes fn with the active transaction,
// and manages the full transaction lifecycle. If fn returns an error, the
// transaction is rolled back and that error is returned (optionally wrapped
// with any rollback error). If fn completes without error, the transaction is
// committed and the result from fn is returned.
//
// If a panic occurs while fn is executing, a deferred recovery handler rolls
// back the transaction and then re-panics with the original value. This
// ensures that no partial changes are committed while still propagating the
// panic to the caller.
func WithTx[T any](ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) (*T, error)) (*T, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	res, err := fn(tx)
	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %w", err, rerr)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return res, nil
}
