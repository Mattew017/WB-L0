package usecase

import (
	"context"
	"log"
	"web_service/internal/domain"
	"web_service/internal/protocols"
)

type SaveOrderUseCase struct {
	orderRepo    protocols.OrderRepoInterface
	paymentRepo  protocols.PaymentRepoInterface
	deliveryRepo protocols.DeliveryRepoInterface
	itemRepo     protocols.ItemRepoInterface
	txManager    protocols.TransactionManagerInterface
}

func NewSaveOrderUseCase(
	orderRepo protocols.OrderRepoInterface,
	paymentRepo protocols.PaymentRepoInterface,
	deliveryRepo protocols.DeliveryRepoInterface,
	itemRepo protocols.ItemRepoInterface,
	txManager protocols.TransactionManagerInterface,
) *SaveOrderUseCase {
	return &SaveOrderUseCase{orderRepo: orderRepo, paymentRepo: paymentRepo,
		deliveryRepo: deliveryRepo, itemRepo: itemRepo,
		txManager: txManager}
}

func (uc *SaveOrderUseCase) Save(order *domain.Order) error {
	ctx := context.Background()
	err := uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		existingOrder, err := uc.orderRepo.GetById(ctx, order.OrderUID)
		if existingOrder != nil {
			log.Printf("order %s already exists: %s", order.OrderUID, err)
			return domain.OrderAlreadyExistsError
		}
		err = uc.orderRepo.Save(ctx, order)
		if err != nil {
			log.Println(err)
			return err
		}
		order.Delivery.OrderUID = order.OrderUID
		err = uc.deliveryRepo.Save(ctx, &order.Delivery)
		if err != nil {
			log.Println(err)
			return err
		}
		order.Payment.Transaction = order.OrderUID
		err = uc.paymentRepo.Save(ctx, &order.Payment)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range order.Items {
			err = uc.itemRepo.Save(ctx, &item)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		return err
	})
	if err != nil {
		return err
	}

	return nil
}
