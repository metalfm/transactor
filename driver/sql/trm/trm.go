package trm

import (
	"context"
	"database/sql"
	"fmt"
)

type impl[T any] struct {
	db *sql.DB
	wt withTx[T]
}

func New[T withTx[T]](db *sql.DB, wt T) *impl[T] {
	return &impl[T]{
		db: db,
		wt: wt,
	}
}

func (slf *impl[T]) InTx(
	ctx context.Context,
	fn func(repo T) error,
) error {
	tx, err := slf.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

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
