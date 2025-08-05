package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"testberry/internal/domain/service"
	"testberry/internal/ports"
)

type Handler struct {
	service *service.Service
	logger  ports.Logger
}

func NewHandler(service *service.Service, logger ports.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderUID := strings.TrimPrefix(r.URL.Path, "/order/")
	h.logger.Info("orderUID")
	order, err := h.service.GetOrder(r.Context(), orderUID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
