package app

import (
	"context"
	"fmt"
	"github.com/metalfm/transactor/tr"
)

type Service[T repoTx] struct {
	tr tr.Transactor[T]
}

func NewService[T repoTx](tr tr.Transactor[T]) *Service[T] {
	return &Service[T]{tr}
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
