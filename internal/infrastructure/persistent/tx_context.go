package persistent

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type querierKey struct{}

// Querier — интерфейс, общий для *pgxpool.Pool и pgx.Tx
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func ContextWithQuerier(ctx context.Context, q Querier) context.Context {
	return context.WithValue(ctx, querierKey{}, q)
}

func QuerierFromContext(ctx context.Context) Querier {
	if q, ok := ctx.Value(querierKey{}).(Querier); ok {
		return q
	}
	return nil
}
