package ports

import (
	"context"
	order_entity "testberry/internal/domain/order"
)

type Repository interface {
	SaveOrder(ctx context.Context, order order_entity.Order) error
	GetOrderByID(ctx context.Context, orderUID string) (order_entity.Order, error)
	RestoreCache(ctx context.Context) ([]order_entity.Order, error)
}
