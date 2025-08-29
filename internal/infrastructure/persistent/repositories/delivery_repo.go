package repositories

import (
	"context"
	"fmt"
	"log"
	"web_service/internal/domain"
	"web_service/internal/infrastructure/persistent"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryRepo struct {
	pool *pgxpool.Pool
}

func NewDeliveryRepo(pool *pgxpool.Pool) *DeliveryRepo {
	return &DeliveryRepo{pool: pool}
}

func (r *DeliveryRepo) getQuerier(ctx context.Context) persistent.Querier {
	if q := persistent.QuerierFromContext(ctx); q != nil {
		return q
	}
	return r.pool
}

func (r *DeliveryRepo) GetByOrderId(ctx context.Context, orderUid string) (*domain.Delivery, error) {
	querier := r.getQuerier(ctx)

	row := querier.QueryRow(ctx,
		`SELECT order_uid, name, phone, zip, city, address, region, email
         FROM deliveries 
         WHERE order_uid = $1`,
		orderUid,
	)

	var delivery domain.Delivery
	err := row.Scan(
		&delivery.OrderUID,
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("get payment database error: %w", err)
	}

	return &delivery, nil
}

func (r *DeliveryRepo) Save(ctx context.Context, delivery *domain.Delivery) error {
	querier := r.getQuerier(ctx)

	_, err := querier.Exec(ctx,
		`INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		delivery.OrderUID,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	)
	if err != nil {
		log.Printf("Save delivery error: %v", err)
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	return nil
}
