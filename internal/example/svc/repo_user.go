package svc

import (
	"context"
	"fmt"
	"github.com/metalfm/transactor/driver/sql/trm"
	"github.com/metalfm/transactor/internal/example/app"
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

func (slf *RepoUser) FindUserByID(ctx context.Context, id int64) (app.User, error) {
	var user app.User

	query := "SELECT id, name FROM users WHERE id = $1"

	err := slf.q.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name)
	if err != nil {
		return app.User{}, fmt.Errorf("find user by id='%d': %w", id, err)
	}

	return user, nil
}

func (slf *RepoUser) CreateUser(ctx context.Context, name string) error {
	query := "INSERT INTO users (name) VALUES ($1)"

	_, err := slf.q.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}
