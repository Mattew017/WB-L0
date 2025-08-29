package http_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"web_service/internal/domain"
	"web_service/internal/usecase"
)

type GetOrderHandler struct {
	useCase *usecase.GetOrderUseCase
}

func NewOrderHandler(useCase *usecase.GetOrderUseCase) *GetOrderHandler {
	return &GetOrderHandler{useCase: useCase}
}

func (h *GetOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	// Извлекаем order_uid из URL пути
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 || pathParts[2] == "" {
		http.Error(w, `{"error": "Order ID is required"}`, http.StatusBadRequest)
		return
	}
	orderUID := pathParts[2]
	if orderUID == "" {
		http.Error(w, `{"error": "order_uid parameter is required"}`, http.StatusBadRequest)
		return
	}
	order, err := h.useCase.GetOrderById(orderUID)
	if err != nil {
		if errors.Is(err, domain.OrderNotFoundError) {
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
