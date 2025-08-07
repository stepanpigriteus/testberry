package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	order_entity "testberry/internal/domain/order"
	testmock "testberry/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_GetOrder(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		setupMock      func(*testmock.MockOrderService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Успешное получение заказа",
			url:  "/order/12345678901234567890",
			setupMock: func(mockService *testmock.MockOrderService) {
				mockService.On("GetOrder", mock.Anything, "12345678901234567890").Return(testmock.Test_order, nil)
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
			setupMock: func(mockService *testmock.MockOrderService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing order UID\n",
		},
		{
			name: "Некорректная длина UID заказа",
			url:  "/order/123",
			setupMock: func(mockService *testmock.MockOrderService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid order UID lenght\n",
		},
		{
			name: "Ошибка сервиса при получении заказа",
			url:  "/order/12345678901234567890",
			setupMock: func(mockService *testmock.MockOrderService) {
				mockService.On("GetOrder", mock.Anything, "12345678901234567890").Return(order_entity.Order{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error\n",
		},
		{
			name: "Пустой путь после /order/",
			url:  "/order/",
			setupMock: func(mockService *testmock.MockOrderService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing order UID\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(testmock.MockOrderService)
			testLogger := &testmock.TestLogger{}
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
	mockService := new(testmock.MockOrderService)
	testLogger := &testmock.TestLogger{}

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
	mockService := new(testmock.MockOrderService)
	testLogger := &testmock.TestLogger{}

	handler := NewHandler(mockService, testLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
	assert.Equal(t, testLogger, handler.logger)
}
