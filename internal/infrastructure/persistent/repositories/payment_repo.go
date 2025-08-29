package repositories

import (
	"context"
	"fmt"
	"log"
	"web_service/internal/domain"
	"web_service/internal/infrastructure/persistent"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepo struct {
	pool *pgxpool.Pool
}

func (r *PaymentRepo) getQuerier(ctx context.Context) persistent.Querier {
	if q := persistent.QuerierFromContext(ctx); q != nil {
		return q
	}
	return r.pool
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

func (r *PaymentRepo) GetByTransactionId(ctx context.Context, transaction string) (*domain.Payment, error) {
	querier := r.getQuerier(ctx)
	row := querier.QueryRow(
		ctx,
		"select id, transaction, request_id, currency, provider, amount, payment_dt, bank,"+
			"delivery_cost, goods_total, custom_fee"+
			" from payments where transaction = $1",
		transaction)
	var payment domain.Payment
	err := row.Scan(&payment.ID, &payment.Transaction, &payment.RequestID,
		&payment.Currency, &payment.Provider,
		&payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost,
		&payment.GoodsTotal, &payment.CustomFee)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &payment, nil
}

func (r *PaymentRepo) Save(ctx context.Context, payment *domain.Payment) error {
	querier := r.getQuerier(ctx)

	_, err := querier.Exec(ctx,
		`INSERT INTO payments (
            transaction, request_id, currency, provider, amount, payment_dt, bank, 
            delivery_cost, goods_total, custom_fee
         ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		payment.Transaction,
		payment.RequestID,
		payment.Currency,
		payment.Provider,
		payment.Amount,
		payment.PaymentDt,
		payment.Bank,
		payment.DeliveryCost,
		payment.GoodsTotal,
		payment.CustomFee,
	)
	if err != nil {
		log.Printf("Save payment error: %v\n", err)
		return fmt.Errorf("failed to save payment: %w", err)
	}

	return nil
}
