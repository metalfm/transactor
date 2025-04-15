package benchmark_test

import (
	"context"
	"database/sql"
	"github.com/aneshas/tx/v2/sqltx"
	_ "github.com/lib/pq"
	"github.com/metalfm/transactor/driver/sql/trm"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	avito "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	aneshas "github.com/aneshas/tx/v2"
)

func BenchmarkSQLPostgres(b *testing.B) {
	b.Run("tx=native", func(b *testing.B) {
		ctx := context.Background()

		conn, cleanup := prepare(ctx, b)
		defer cleanup()

		r := &repo{}

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tx, err := conn.BeginTx(ctx, nil)
				require.NoError(b, err)

				err = r.CreateNative(ctx, tx, "some user name")
				require.NoError(b, err)

				err = tx.Commit()
				require.NoError(b, err)
			}
		})
	})
	b.Run("tx=transactor", func(b *testing.B) {
		ctx := context.Background()

		conn, cleanup := prepare(ctx, b)
		defer cleanup()

		r := repo{db0: conn}
		tr := trm.New(conn, &r)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := tr.InTx(ctx, func(repo *repo) error {
					return repo.CreateTransactor(ctx, "some user name")
				})
				require.NoError(b, err)
			}
		})
	})
	b.Run("tx=avito", func(b *testing.B) {
		ctx := context.Background()

		conn, cleanup := prepare(ctx, b)
		defer cleanup()

		r := repo{db1: conn, getter: avito.DefaultCtxGetter}
		trManager := manager.Must(avito.NewDefaultFactory(conn))

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := trManager.Do(ctx, func(ctx context.Context) error {
					return r.CreateAvito(ctx, "some user name")
				})
				require.NoError(b, err)
			}
		})
	})
	b.Run("tx=aneshas", func(b *testing.B) {
		ctx := context.Background()

		conn, cleanup := prepare(ctx, b)
		defer cleanup()

		r := repo{db1: conn}
		trManager := aneshas.New(sqltx.NewDB(conn))

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := trManager.WithTransaction(ctx, func(ctx context.Context) error {
					return r.CreateAneshas(ctx, "some user name")
				})
				require.NoError(b, err)
			}
		})
	})
}

type repo struct {
	db0    trm.Query
	db1    *sql.DB
	getter *avito.CtxGetter
}

func (slf *repo) CreateNative(ctx context.Context, tx *sql.Tx, name string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO users (name) VALUES ($1)`, name)
	return err
}

func (slf *repo) WithTx(tx trm.Transaction) *repo {
	return &repo{db0: tx}
}

func (slf *repo) CreateTransactor(ctx context.Context, name string) error {
	_, err := slf.db0.ExecContext(ctx, `INSERT INTO users (name) VALUES ($1)`, name)
	return err
}

func (slf *repo) CreateAvito(ctx context.Context, name string) error {
	_, err := slf.getter.
		DefaultTrOrDB(ctx, slf.db1).
		ExecContext(ctx, `INSERT INTO users (name) VALUES ($1)`, name)
	return err
}

func (slf *repo) CreateAneshas(ctx context.Context, name string) error {
	_, err := slf.conn(ctx).ExecContext(ctx, `INSERT INTO users (name) VALUES ($1)`, name)
	return err
}

func (slf *repo) conn(ctx context.Context) trm.Query {
	if tx, ok := sqltx.From(ctx); ok {
		return tx
	}

	return slf.db1
}

func prepare(ctx context.Context, tb testing.TB) (*sql.DB, func()) {
	conn, err := sql.Open("postgres", os.Getenv("DSN_POSTGRES"))
	require.NoError(tb, err)

	err = conn.Ping()
	require.NoError(tb, err)

	createSQL := `CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT NOT NULL)`

	_, err = conn.ExecContext(ctx, createSQL)
	require.NoError(tb, err)

	_, err = conn.ExecContext(ctx, "DELETE FROM USERS")
	require.NoError(tb, err)

	return conn, func() {
		_, err = conn.ExecContext(ctx, "DROP TABLE users")
		require.NoError(tb, err)

		err = conn.Close()
		require.NoError(tb, err)
	}
}
