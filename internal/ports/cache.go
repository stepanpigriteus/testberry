package ports

import (
	"context"
	order_entity "testberry/internal/domain/order"
)

type Cache interface {
	Set(ctx context.Context, order order_entity.Order) error
	Get(ctx context.Context, orderUID string) (order_entity.Order, bool, error)
}
