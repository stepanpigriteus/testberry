package cache

import (
	"context"
	"encoding/json"

	order_entity "testberry/internal/domain/order"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
}

func NewCache(addr, password string, db int) *Cache {
	return &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (c *Cache) Set(ctx context.Context, order order_entity.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, order.OrderUID, data, 0).Err()
}

func (c *Cache) Get(ctx context.Context, orderUID string) (order_entity.Order, bool, error) {
	val, err := c.client.Get(ctx, orderUID).Result()
	if err == redis.Nil {
		return order_entity.Order{}, false, nil
	}
	if err != nil {
		return order_entity.Order{}, false, err
	}
	var order order_entity.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return order_entity.Order{}, false, err
	}
	return order, true, nil
}
