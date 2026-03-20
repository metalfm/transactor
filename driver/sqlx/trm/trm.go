package trm

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type impl[T any] struct {
	db *sqlx.DB
	wt withTx[T]
}

//nolint:revive // exported constructor intentionally returns hidden implementation type
func New[T withTx[T]](db *sqlx.DB, wt T) *impl[T] {
	return &impl[T]{
		db: db,
		wt: wt,
	}
}

func (slf *impl[T]) InTx(
	ctx context.Context,
	fn func(repo T) error,
) error {
	tx, err := slf.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	err = fn(slf.wt.WithTx(tx))
	if err != nil {
		return fmt.Errorf("trm callback: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

var _ Transaction = (*sqlx.Tx)(nil)
