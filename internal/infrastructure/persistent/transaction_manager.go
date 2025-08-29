package persistent

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxTransactionManager struct {
	pool *pgxpool.Pool
}

func NewPgxTransactionManager(pool *pgxpool.Pool) *PgxTransactionManager {
	return &PgxTransactionManager{pool: pool}
}

func (tm *PgxTransactionManager) WithinTransaction(ctx context.Context,
	fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrTxClosed) {
			return
		}
	}(tx, ctx)

	// Устанавливаем tx в контекст как Querier
	ctx = ContextWithQuerier(ctx, tx)

	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
