package repositories

import (
	"context"
	"fmt"
	"log"
	"web_service/internal/domain"
	"web_service/internal/infrastructure/persistent"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepo struct {
	pool *pgxpool.Pool
}

func NewItemRepo(pool *pgxpool.Pool) *ItemRepo {
	return &ItemRepo{pool: pool}
}

func (r *ItemRepo) getQuerier(ctx context.Context) persistent.Querier {
	if q := persistent.QuerierFromContext(ctx); q != nil {
		return q
	}
	return r.pool
}

func (r *ItemRepo) GetByTrackNumber(ctx context.Context, trackNumber string) ([]domain.Item, error) {
	querier := r.getQuerier(ctx)

	rows, err := querier.Query(ctx,
		`SELECT 
            	rid, track_number, chrt_id, price, name, sale, size,
    			total_price, nm_id, brand, status
			 FROM items 
			 WHERE track_number = $1`,
		trackNumber,
	)
	if err != nil {
		log.Printf("get items database error: %v", err)
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		var item domain.Item
		err := rows.Scan(
			&item.Rid,
			&item.TrackNumber,
			&item.ChrtID,
			&item.Price,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			log.Printf("failed to scan item: %v", err)
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("items Rows iteration error: %v", err)
		return nil, fmt.Errorf("error iterating items rows: %w", err)
	}

	return items, nil
}

func (r *ItemRepo) Save(ctx context.Context, item *domain.Item) error {
	querier := r.getQuerier(ctx)

	_, err := querier.Exec(ctx,
		`INSERT INTO items (
            rid, track_number, chrt_id, price, name, sale, size, 
            total_price, nm_id, brand, status
         ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		item.Rid,
		item.TrackNumber,
		item.ChrtID,
		item.Price,
		item.Name,
		item.Sale,
		item.Size,
		item.TotalPrice,
		item.NmID,
		item.Brand,
		item.Status,
	)
	if err != nil {
		log.Printf("failed to save item: %v", err)
		return fmt.Errorf("failed to save item: %w", err)
	}

	return nil
}
