package svc

import (
	"context"
	"fmt"
	"github.com/metalfm/transactor/driver/sql/trm"
)

type RepoUser struct {
	q trm.Query
}

func NewRepoUser(q trm.Query) *RepoUser {
	return &RepoUser{q}
}

func (slf *RepoUser) WithTx(tx trm.Transaction) *RepoUser {
	return &RepoUser{tx}
}

func (slf *RepoUser) CreateUser(ctx context.Context, name string) error {
	query := "INSERT INTO users (name) VALUES ($1)"

	_, err := slf.q.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}
