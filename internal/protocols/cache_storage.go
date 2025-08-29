package protocols

import "web_service/internal/domain"

type OrderStorageInterface interface {
	Get(orderUID string) (*domain.Order, error)
	Save(orderUID string, order *domain.Order)
}
