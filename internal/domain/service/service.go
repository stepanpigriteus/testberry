package service

import (
	"context"
	"encoding/json"
	order_entity "testberry/internal/domain/order"
	"testberry/internal/ports"
)

type Service struct {
	repo   ports.Repository
	cache  ports.Cache
	broker ports.MessageBroker
	logger ports.Logger
}

func NewService(repo ports.Repository, cache ports.Cache, broker ports.MessageBroker, logger ports.Logger) *Service {
	return &Service{repo: repo, cache: cache, broker: broker, logger: logger}
}

func (s *Service) GetOrder(ctx context.Context, orderUID string) (order_entity.Order, error) {
	s.logger.Info("GetOrderService called")
	if order, exists, err := s.cache.Get(ctx, orderUID); err == nil && exists {
		return order, nil
	}
	return s.repo.GetOrderByID(ctx, orderUID)
}

func (s *Service) Start(ctx context.Context) error {
	orders, err := s.repo.RestoreCache(ctx)
	if err != nil {
		return err
	}
	for _, order := range orders {
		if err := s.cache.Set(ctx, order); err != nil {
			return err
		}
	}

	return s.broker.Consume(ctx, func(ctx context.Context, message []byte) error {
		var order order_entity.Order
		if err := json.Unmarshal(message, &order); err != nil {
			s.logger.Error("Message unmarshall failed")
			return nil
		}
		if err := s.repo.SaveOrder(ctx, order); err != nil {
			return err
		}
		return s.cache.Set(ctx, order)
	})
}
