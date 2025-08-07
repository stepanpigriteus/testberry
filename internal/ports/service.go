package ports

import (
	"context"
	order_entity "testberry/internal/domain/order"
)

type OrderService interface {
	GetOrder(ctx context.Context, orderUID string) (order_entity.Order, error)
}
