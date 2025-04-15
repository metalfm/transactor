package tr

import (
	"context"
)

type Transactor[T any] interface {
	InTx(ctx context.Context, fn func(T) error) error
}
