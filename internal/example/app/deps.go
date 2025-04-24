package app

import "context"

//go:generate go tool mockgen -typed -source=$GOFILE -destination=./mock/$GOFILE

type repoTx interface {
	CreateUser(ctx context.Context, name string) error
	CreateOrder(ctx context.Context, items []string) error
}
