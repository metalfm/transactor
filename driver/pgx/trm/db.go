package trm

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Query interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Transaction interface {
	Query
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type db interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

type withTx[T any] interface {
	WithTx(tx Transaction) T
}
