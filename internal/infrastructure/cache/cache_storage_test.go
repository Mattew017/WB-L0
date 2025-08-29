package cache

import (
	"strconv"
	"sync"
	"testing"
	"web_service/internal/domain"
	"web_service/internal/protocols"

	"github.com/stretchr/testify/assert"
)

func TestPutAndGet(t *testing.T) {
	var storage protocols.OrderStorageInterface
	storage = NewLocalOrderStorage()

	orderUID := "test"
	order := &domain.Order{OrderUID: orderUID}
	storage.Save(order.OrderUID, order)
	_, err := storage.Get(orderUID)
	if err != nil {
		t.Errorf("Error getting order from storage: %v", err)
	}
}

func TestPutAndGetAsync(t *testing.T) {
	var storage protocols.OrderStorageInterface
	storage = NewLocalOrderStorage()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			id := strconv.Itoa(i)
			newOrder := &domain.Order{OrderUID: id}
			storage.Save(id, newOrder)
		}
		wg.Done()
	}()
	wg.Wait()
	for i := 0; i < 100; i++ {
		id := strconv.Itoa(i)
		actual, ok := storage.Get(id)
		assert.Equal(t, nil, ok)
		assert.Equal(t, id, actual.OrderUID)
	}
}
