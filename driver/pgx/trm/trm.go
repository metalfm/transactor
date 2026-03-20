package trm

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type impl[T any] struct {
	db db
	wt withTx[T]
}

//nolint:revive // exported constructor intentionally returns hidden implementation type
func New[T withTx[T]](db db, wt T) *impl[T] {
	return &impl[T]{
		db: db,
		wt: wt,
	}
}

func (slf *impl[T]) InTx(
	ctx context.Context,
	fn func(repo T) error,
) error {
	tx, err := slf.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = fn(slf.wt.WithTx(tx))
	if err != nil {
		return fmt.Errorf("trm callback: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
