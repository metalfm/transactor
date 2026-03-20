package trm

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Query interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
}

type Transaction interface {
	Query
	Commit() error
	Rollback() error
}

type withTx[T any] interface {
	WithTx(tx Transaction) T
}
