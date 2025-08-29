package cache

import (
	"log"
	"sync"
	"web_service/internal/domain"
)

type LocalOrderStorage struct {
	data sync.Map
}

func NewLocalOrderStorage() *LocalOrderStorage {
	return &LocalOrderStorage{}
}

func (s *LocalOrderStorage) Get(orderUID string) (*domain.Order, error) {
	val, ok := s.data.Load(orderUID)
	if !ok {
		log.Printf("order %s not found in cache\n", orderUID)
		return nil, domain.OrderNotFoundError
	}
	order, ok := val.(*domain.Order)
	if !ok {
		log.Printf("order %s not found in cache: could not serialize to domain entity\n", orderUID)
		return nil, domain.OrderNotFoundError
	}
	log.Printf("found order %s in cache\n", orderUID)
	return order, nil
}

func (s *LocalOrderStorage) Save(orderUID string, order *domain.Order) {
	s.data.Store(orderUID, order)
	log.Printf("saved order %s in cache\n", orderUID)
}
