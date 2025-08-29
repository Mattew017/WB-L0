package repositories

import (
	"context"
	"fmt"
	"web_service/internal/domain"
	"web_service/internal/infrastructure/persistent"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) getQuerier(ctx context.Context) persistent.Querier {
	if q := persistent.QuerierFromContext(ctx); q != nil {
		return q
	}
	return r.pool
}

func (r *OrderRepo) GetById(ctx context.Context, orderUid string) (*domain.Order, error) {
	querier := r.getQuerier(ctx)
	row := querier.QueryRow(
		ctx,
		`select order_uid, track_number, entry, locale, internal_signature,
				customer_id, delivery_service, shardkey, sm_id, date_created,
				oof_shard from orders where order_uid = $1`,
		orderUid)
	var order domain.Order
	err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry,
		&order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey,
		&order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("%w: database has no order with uid %s", domain.OrderNotFoundError, orderUid)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &order, nil
}

func (r *OrderRepo) Save(ctx context.Context, order *domain.Order) error {
	querier := r.getQuerier(ctx)
	_, err := querier.Exec(ctx,
		`INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
         ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry,
		order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey,
		order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	return nil

}

func (r *OrderRepo) GetAllOrderUIDs(ctx context.Context) ([]string, error) {
	querier := r.getQuerier(ctx)
	rows, err := querier.Query(ctx, `SELECT order_uid FROM orders`)
	if err != nil {
		return nil, fmt.Errorf("could not load all order uids, database error: %w", err)
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("could not scan order uid row: %w", err)
		}
		result = append(result, orderUID)
	}
	return result, nil
}
