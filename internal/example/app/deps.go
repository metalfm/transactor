package app

import "context"

//go:generate go tool mockgen -typed -source=$GOFILE -destination=./mock/$GOFILE

type repo interface {
	FindUserByID(ctx context.Context, id int64) (User, error)
}

type repoTx interface {
	CreateUser(ctx context.Context, name string) error
	CreateOrder(ctx context.Context, items []string) error
}
