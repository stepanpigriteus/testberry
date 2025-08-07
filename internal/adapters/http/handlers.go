package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"testberry/internal/ports"
)

type Handler struct {
	service ports.OrderService
	logger  ports.Logger
}

func NewHandler(service ports.OrderService, logger ports.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	orderUID := strings.TrimPrefix(r.URL.Path, "/order/")
	if orderUID == "" {
		http.Error(w, "Missing order UID", http.StatusBadRequest)
		return
	}
	if len(orderUID) != 20 {
		http.Error(w, "Invalid order UID lenght", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrder(r.Context(), orderUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(order); err != nil {
		h.logger.Error("failed to encode order to JSON", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
