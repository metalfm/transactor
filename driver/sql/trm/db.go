package trm

import (
	"context"
	"database/sql"
)

type Query interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Transaction interface {
	Query
	Commit() error
	Rollback() error
}

type withTx[T any] interface {
	WithTx(tx Transaction) T
}
