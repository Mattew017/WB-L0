package protocols

import (
	"context"
	"web_service/internal/domain"
)

type OrderRepoInterface interface {
	GetById(ctx context.Context, orderUid string) (*domain.Order, error)
	GetAllOrderUIDs(ctx context.Context) ([]string, error)
	Save(ctx context.Context, order *domain.Order) error
}

type DeliveryRepoInterface interface {
	GetByOrderId(ctx context.Context, orderUid string) (*domain.Delivery, error)
	Save(ctx context.Context, delivery *domain.Delivery) error
}

type PaymentRepoInterface interface {
	GetByTransactionId(ctx context.Context, transaction string) (*domain.Payment, error)
	Save(ctx context.Context, payment *domain.Payment) error
}

type ItemRepoInterface interface {
	GetByTrackNumber(ctx context.Context, trackNumber string) ([]domain.Item, error)
	Save(ctx context.Context, item *domain.Item) error
}
