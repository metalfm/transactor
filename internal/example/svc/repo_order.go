package svc

import (
	"context"
	"fmt"
	"github.com/metalfm/transactor/driver/sql/trm"
)

type RepoOrder struct {
	q trm.Query
}

func NewRepoOrder(q trm.Query) *RepoOrder {
	return &RepoOrder{q}
}

func (slf *RepoOrder) WithTx(tx trm.Transaction) *RepoOrder {
	return &RepoOrder{tx}
}

func (slf *RepoOrder) CreateOrder(ctx context.Context, items []string) error {
	query := "INSERT INTO orders (item) VALUES ($1)"

	for _, item := range items {
		_, err := slf.q.ExecContext(ctx, query, item)
		if err != nil {
			return fmt.Errorf("create order for item '%s': %w", item, err)
		}
	}

	return nil
}
