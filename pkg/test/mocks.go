package testmock

import (
	"context"
	order_entity "testberry/internal/domain/order"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) GetOrder(ctx context.Context, orderUID string) (order_entity.Order, error) {
	args := m.Called(ctx, orderUID)
	return args.Get(0).(order_entity.Order), args.Error(1)
}

func (m *MockOrderService) SaveOrder(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOrderService) SendRandomOrder(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOrderService) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetOrderByID(ctx context.Context, uid string) (order_entity.Order, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(order_entity.Order), args.Error(1)
}

func (m *MockRepository) RestoreCache(ctx context.Context) ([]order_entity.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]order_entity.Order), args.Error(1)
}

func (m *MockRepository) SaveOrder(ctx context.Context, order order_entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, uid string) (order_entity.Order, bool, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(order_entity.Order), args.Bool(1), args.Error(2)
}

func (m *MockCache) Set(ctx context.Context, o order_entity.Order) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}

type MockConsumer struct {
	ConsumeFunc func(ctx context.Context, handler func(context.Context, []byte) error) error
}

func (m *MockConsumer) Consume(ctx context.Context, handler func(context.Context, []byte) error) error {
	return m.ConsumeFunc(ctx, handler)
}

type TestLogger struct{}

func (l *TestLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *TestLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *TestLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *TestLogger) Debug(msg string, keysAndValues ...interface{}) {}

var Test_order order_entity.Order = order_entity.Order{
	OrderUID:        "12345678901234567890",
	TrackNumber:     "WBILMTESTTRACK",
	Entry:           "WBIL",
	CustomerID:      "test",
	Locale:          "en",
	Shardkey:        "9",
	SmID:            99,
	DateCreated:     time.Now(),
	OofShard:        "1",
	DeliveryService: "DHL",
	Delivery: order_entity.Delivery{
		Name:    "Test User",
		Phone:   "+79001234567",
		Zip:     "2639809",
		City:    "Kiryat Mozkin",
		Address: "Ploshad Mira 15",
		Region:  "Kraiot",
		Email:   "test@gmail.com",
	},
	Payment: order_entity.Payment{
		Transaction:  "b563feb7b2b84b6test",
		RequestID:    "req123",
		Currency:     "USD",
		Provider:     "wbpay",
		Amount:       1817,
		PaymentDt:    1637907727,
		Bank:         "alpha",
		DeliveryCost: 1500,
		GoodsTotal:   317,
		CustomFee:    0,
	},
	Items: []order_entity.Item{
		{
			ChrtID:      9934930,
			TrackNumber: "WBILMTESTTRACK",
			Price:       453,
			Rid:         "ab4219087a764ae0btest",
			Name:        "Mascaras",
			Sale:        30,
			Size:        "0",
			TotalPrice:  317,
			NmID:        2389212,
			Brand:       "Vivienne Sabo",
			Status:      202,
		},
	},
}
