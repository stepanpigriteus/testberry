package service

import (
	"context"
	"errors"
	"fmt"

	order_entity "testberry/internal/domain/order"
	testmock "testberry/pkg/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetOrder_FromCache(t *testing.T) {
	ctx := context.Background()
	orderUID := "12345678901234567890"

	mockRepo := new(testmock.MockRepository)
	mockCache := new(testmock.MockCache)
	mockLogger := &testmock.TestLogger{}

	mockCache.On("Get", ctx, orderUID).Return(testmock.Test_order, true, nil)
	s := &Service{
		repo:   mockRepo,
		cache:  mockCache,
		logger: mockLogger,
	}
	result, err := s.GetOrder(ctx, orderUID)
	assert.NoError(t, err)
	assert.Equal(t, testmock.Test_order.OrderUID, result.OrderUID)

	mockCache.AssertExpectations(t)
}

func TestService_GetOrder_CachMiss(t *testing.T) {
	ctx := context.Background()
	orderUID := "12345678901234567890"
	expectedOrder := testmock.Test_order

	mockRepo := new(testmock.MockRepository)
	mockCache := new(testmock.MockCache)
	mockLogger := &testmock.TestLogger{}

	mockCache.On("Get", ctx, orderUID).Return(order_entity.Order{}, false, nil)
	mockRepo.On("GetOrderByID", ctx, orderUID).Return(expectedOrder, nil)
	mockCache.On("Set", ctx, expectedOrder).Return(nil)

	s := &Service{
		repo:   mockRepo,
		cache:  mockCache,
		logger: mockLogger,
	}

	result, err := s.GetOrder(ctx, orderUID)
	assert.NoError(t, err)
	assert.Equal(t, testmock.Test_order, result)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestService_GetOrder_CacheFail(t *testing.T) {
	ctx := context.Background()
	orderUID := "1234567890123456S890"

	mockRepo := new(testmock.MockRepository)
	mockCache := new(testmock.MockCache)
	mockLogger := &testmock.TestLogger{}

	mockCache.On("Get", ctx, orderUID).Return(order_entity.Order{}, false, errors.New("cache down"))
	s := &Service{
		repo:   mockRepo,
		cache:  mockCache,
		logger: mockLogger,
	}
	_, err := s.GetOrder(ctx, orderUID)
	assert.Error(t, err)
	assert.EqualError(t, err, "cache down")
	mockCache.AssertExpectations(t)

}

func TestService_GetOrder_Cache_Miss_DB_Error(t *testing.T) {
	ctx := context.Background()
	orderUID := "1234567890123456S890"

	mockRepo := new(testmock.MockRepository)
	mockCache := new(testmock.MockCache)
	mockLogger := &testmock.TestLogger{}

	mockCache.On("Get", ctx, orderUID).Return(order_entity.Order{}, false, nil)
	mockRepo.On("GetOrderByID", ctx, orderUID).Return(order_entity.Order{}, fmt.Errorf("order with UID %s not found", orderUID))
	s := &Service{
		repo:   mockRepo,
		cache:  mockCache,
		logger: mockLogger,
	}
	order, err := s.GetOrder(ctx, orderUID)
	assert.Error(t, err)
	assert.Equal(t, order_entity.Order{}, order)
	mockCache.AssertExpectations(t)

}

func TestService_GetOrder_CacheMiss_SetError(t *testing.T) {
	ctx := context.Background()
	orderUID := "12345678901234567890"
	expectedOrder := testmock.Test_order

	mockRepo := new(testmock.MockRepository)
	mockCache := new(testmock.MockCache)
	mockLogger := &testmock.TestLogger{}

	mockCache.On("Get", ctx, orderUID).Return(order_entity.Order{}, false, nil)
	mockRepo.On("GetOrderByID", ctx, orderUID).Return(expectedOrder, nil)
	mockCache.On("Set", ctx, expectedOrder).Return(errors.New("cache set error"))

	s := &Service{
		repo:   mockRepo,
		cache:  mockCache,
		logger: mockLogger,
	}

	result, err := s.GetOrder(ctx, orderUID)
	assert.Error(t, err)
	assert.EqualError(t, err, "cache set error")
	assert.Equal(t, expectedOrder, result)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)

}
