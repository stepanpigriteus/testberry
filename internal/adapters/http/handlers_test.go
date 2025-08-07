package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	order_entity "testberry/internal/domain/order"

	"github.com/stretchr/testify/assert"
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

type TestLogger struct{}

func (l *TestLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *TestLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *TestLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *TestLogger) Debug(msg string, keysAndValues ...interface{}) {}

func TestHandler_GetOrder(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		setupMock      func(*MockOrderService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Успешное получение заказа",
			url:  "/order/12345678901234567890",
			setupMock: func(mockService *MockOrderService) {
				order := order_entity.Order{
					OrderUID:    "12345678901234567890",
					TrackNumber: "WBILMTESTTRACK",
					Entry:       "WBIL",
					CustomerID:  "test",
					Locale:      "en",
					Shardkey:    "9",
					SmID:        99,
					DateCreated: time.Now(),
					OofShard:    "1",
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
						RequestID:    "",
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
				mockService.On("GetOrder", mock.Anything, "12345678901234567890").Return(order, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

				var order order_entity.Order
				err := json.Unmarshal(w.Body.Bytes(), &order)
				assert.NoError(t, err)
				assert.Equal(t, "12345678901234567890", order.OrderUID)
			},
		},
		{
			name: "Отсутствующий UID заказа",
			url:  "/order/",
			setupMock: func(mockService *MockOrderService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing order UID\n",
		},
		{
			name: "Некорректная длина UID заказа",
			url:  "/order/123",
			setupMock: func(mockService *MockOrderService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid order UID lenght\n",
		},
		{
			name: "Ошибка сервиса при получении заказа",
			url:  "/order/12345678901234567890",
			setupMock: func(mockService *MockOrderService) {
				mockService.On("GetOrder", mock.Anything, "12345678901234567890").Return(order_entity.Order{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error\n",
		},
		{
			name: "Пустой путь после /order/",
			url:  "/order/",
			setupMock: func(mockService *MockOrderService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing order UID\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		
			mockService := new(MockOrderService)
			testLogger := &TestLogger{}
	
			tt.setupMock(mockService)

			handler := NewHandler(mockService, testLogger)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			handler.GetOrder(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetOrder_CORS_Headers(t *testing.T) {
	mockService := new(MockOrderService)
	testLogger := &TestLogger{}

	order := order_entity.Order{
		OrderUID: "12345678901234567890",
	}
	mockService.On("GetOrder", mock.Anything, "12345678901234567890").Return(order, nil)

	handler := NewHandler(mockService, testLogger)

	req := httptest.NewRequest(http.MethodGet, "/order/12345678901234567890", nil)
	w := httptest.NewRecorder()

	handler.GetOrder(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestNewHandler(t *testing.T) {
	mockService := new(MockOrderService)
	testLogger := &TestLogger{}

	handler := NewHandler(mockService, testLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
	assert.Equal(t, testLogger, handler.logger)
}
