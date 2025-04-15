package app

import (
	"context"
	"fmt"
	"github.com/metalfm/transactor/tr"
)

type Service[T repoTx] struct {
	tr   tr.Transactor[T]
	repo repo
}

func NewService[T repoTx](
	tr tr.Transactor[T],
	repo repo,
) *Service[T] {
	return &Service[T]{tr, repo}
}

func (slf *Service[T]) FindUser(ctx context.Context, id int64) (User, error) {
	u, err := slf.repo.FindUserByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("find user by id: %w", err)
	}

	return u, nil
}

func (slf *Service[T]) Create(ctx context.Context, name string, items []string) error {
	err := slf.tr.InTx(ctx, func(r T) error {
		err := r.CreateUser(ctx, name)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		err = r.CreateOrder(ctx, items)
		if err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("create user & order: %w", err)
	}

	return nil
}
