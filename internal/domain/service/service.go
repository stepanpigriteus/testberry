package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testberry/internal/adapters/cache"
	messagebrok "testberry/internal/adapters/message_brok"
	"testberry/internal/adapters/postgres"
	order_entity "testberry/internal/domain/order"
	"testberry/internal/ports"
	"testberry/pkg/generator"
	"time"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	repo      ports.Repository
	cache     ports.Cache
	consumer  ports.Consumer
	producer  ports.Producer
	validator *validator.Validate
	logger    ports.Logger
}

func NewService(repo *postgres.Repository, cache *cache.Cache, consumer *messagebrok.Consumer, producer *messagebrok.Producer, logger ports.Logger) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		consumer:  consumer,
		producer:  producer,
		validator: validator.New(),
		logger:    logger,
	}
}

func (s *Service) GetOrder(ctx context.Context, orderUID string) (order_entity.Order, error) {
	s.logger.Info("GetOrderService called")
	order, exists, err := s.cache.Get(ctx, orderUID)
	if err != nil {
		return order, err
	}
	if exists {
		return order, nil
	}
	order, err = s.repo.GetOrderByID(ctx, orderUID)
	if err != nil {
		s.logger.Error("Failed GetOrderService(err in GerOrderRepo)", err)
		return order, err
	}
	err = s.cache.Set(ctx, order)
	if err != nil {
		s.logger.Error("Error setting the value in the cache", err)
		return order, err
	}

	return order, nil
}

func (s *Service) SaveOrder(ctx context.Context) error {
	messageHandler := func(ctx context.Context, message []byte) error {
		var order order_entity.Order
		if err := json.Unmarshal(message, &order); err != nil {
			s.logger.Error("Failed to unmarshal order message:", err)
			return err
		}

		if err := s.validator.Struct(order); err != nil {
			s.logger.Error("Order isn't valid:", err)
			return err
		}
		if err := s.repo.SaveOrder(ctx, order); err != nil {
			s.logger.Error("Order not saved to database:", err)
			return err
		}
		if err := s.cache.Set(ctx, order); err != nil {
			s.logger.Error("Order not saved to cache:", err)
		}

		s.logger.Info("Order successfully processed:", order.OrderUID)
		return nil
	}

	return s.consumer.Consume(ctx, messageHandler)
}

func (s *Service) SendRandomOrder(ctx context.Context) error {
	order := s.generateRandomOrder()
	orderJSON, err := json.Marshal(order)
	if err != nil {
		s.logger.Error("Failed to marshal order:", err)
		return fmt.Errorf("failed to marshal order: %w", err)
	}
	if err := s.producer.Send(order.OrderUID, orderJSON); err != nil {
		s.logger.Error("Failed to send order to producer:", err)
		return fmt.Errorf("failed to send order to producer: %w", err)
	}

	s.logger.Info("Order sent successfully:", order.OrderUID)
	return nil
}

func (s *Service) generateRandomOrder() order_entity.Order {
	return generator.GenerateRandomOrder(time.Now().UnixNano())
}

func (s *Service) Start(ctx context.Context) error {
	orders, err := s.repo.RestoreCache(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if err := s.cache.Set(ctx, order); err != nil {
			s.logger.Error("Failed to restore order to cache:", err)
		}
	}

	s.logger.Info("Cache restored successfully!")
	return nil
}
