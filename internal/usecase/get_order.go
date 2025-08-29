package usecase

import (
	"context"
	"fmt"
	"log"
	"web_service/internal/domain"
	"web_service/internal/protocols"
)

type GetOrderUseCase struct {
	orderRepo    protocols.OrderRepoInterface
	paymentRepo  protocols.PaymentRepoInterface
	deliveryRepo protocols.DeliveryRepoInterface
	itemRepo     protocols.ItemRepoInterface
	txManager    protocols.TransactionManagerInterface
	storage      protocols.OrderStorageInterface
}

func NewGetOrderUseCase(
	orderRepo protocols.OrderRepoInterface,
	paymentRepo protocols.PaymentRepoInterface,
	deliveryRepo protocols.DeliveryRepoInterface,
	itemRepo protocols.ItemRepoInterface,
	txManager protocols.TransactionManagerInterface,
	storage protocols.OrderStorageInterface,
) *GetOrderUseCase {
	return &GetOrderUseCase{orderRepo: orderRepo, paymentRepo: paymentRepo,
		deliveryRepo: deliveryRepo, itemRepo: itemRepo,
		txManager: txManager, storage: storage}
}

func (uc *GetOrderUseCase) GetOrderById(orderUid string) (*domain.Order, error) {
	ctx := context.Background()
	order, err := uc.storage.Get(orderUid)
	if err == nil {
		return order, nil
	}
	order, err = uc.orderRepo.GetById(ctx, orderUid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	delivery, err := uc.deliveryRepo.GetByOrderId(ctx, orderUid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	payment, err := uc.paymentRepo.GetByTransactionId(ctx, orderUid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	items, err := uc.itemRepo.GetByTrackNumber(ctx, order.TrackNumber)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	order.Delivery = *delivery
	order.Payment = *payment
	order.Items = items

	uc.storage.Save(orderUid, order)

	return order, nil
}

func (uc *GetOrderUseCase) RestoreCache(ctx context.Context) error {
	log.Println("Starting cache restoration from database...")
	orderUIDs, err := uc.orderRepo.GetAllOrderUIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get order UIDs: %w", err)
	}
	log.Printf("Found %d orders in database\n", len(orderUIDs))

	successCount := 0
	for _, orderUID := range orderUIDs {
		_, err := uc.GetOrderById(orderUID)
		if err != nil {
			log.Printf("Failed to load order %s to cache: %v\n", orderUID, err)
			continue
		}
		//uc.storage.Save(orderUID, order)
		successCount++
	}
	log.Printf("Cache restoration completed. Loaded %d/%d orders\n", successCount, len(orderUIDs))
	return nil
}
