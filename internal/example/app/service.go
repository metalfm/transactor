package app

import (
	"context"
	"fmt"

	"github.com/metalfm/transactor/tr"
)

type Service struct {
	inTx inTx
}

func NewService[T repoTx](tr tr.Transactor[T]) *Service {
	return &Service{
		inTx: func(ctx context.Context, fn func(repoTx) error) error {
			return tr.InTx(ctx, func(r T) error { return fn(r) })
		},
	}
}

func (slf *Service) Create(ctx context.Context, name string, items []string) error {
	err := slf.inTx(ctx, func(r repoTx) error {
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
