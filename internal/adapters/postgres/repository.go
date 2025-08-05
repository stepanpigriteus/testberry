package postgres

import (
	"context"
	"database/sql"
	order_entity "testberry/internal/domain/order"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveOrder(ctx context.Context, order order_entity.Order) error {
	return nil
}

func (r *Repository) GetOrderByID(ctx context.Context, orderUID string) (order_entity.Order, error) {
	var order order_entity.Order
	return order, nil
}

func (r *Repository) RestoreCache(ctx context.Context) ([]order_entity.Order, error) {

	return nil, nil
}
