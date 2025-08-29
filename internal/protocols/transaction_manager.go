package protocols

import "context"

type TransactionManagerInterface interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
