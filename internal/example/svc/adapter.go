package svc

import (
	"context"
	"github.com/metalfm/transactor/driver/sql/trm"
)

type Adapter struct {
	repoUser  *RepoUser
	repoOrder *RepoOrder
}

func NewAdapter(
	repoUser *RepoUser,
	repoOrder *RepoOrder,
) *Adapter {
	return &Adapter{
		repoUser:  repoUser,
		repoOrder: repoOrder,
	}
}

func (slf *Adapter) WithTx(tx trm.Transaction) *Adapter {
	return &Adapter{
		repoUser:  slf.repoUser.WithTx(tx),
		repoOrder: slf.repoOrder.WithTx(tx),
	}
}

func (slf *Adapter) CreateUser(ctx context.Context, name string) error {
	return slf.repoUser.CreateUser(ctx, name)
}

func (slf *Adapter) CreateOrder(ctx context.Context, items []string) error {
	return slf.repoOrder.CreateOrder(ctx, items)
}
