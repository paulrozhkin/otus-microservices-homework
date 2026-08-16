package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpmiddleware"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/repositories"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type idempotencyRecord struct {
	hash  string
	order *entity.Order
}

type idempotentOrderRepository struct {
	records map[string]idempotencyRecord
	creates int
}

func (r *idempotentOrderRepository) CreateIdempotent(
	_ context.Context,
	order *entity.Order,
	_ *outbox.Message,
	request repositories.IdempotencyRequest,
) (*entity.Order, error) {
	key := fmt.Sprintf("%d:%s", request.UserID, request.Key)
	if record, exists := r.records[key]; exists {
		if record.hash != request.RequestHash {
			return nil, fmt.Errorf("%w: Idempotency-Key was already used with another payload", apperror.ErrAlreadyExists)
		}
		copy := *record.order
		return &copy, nil
	}
	copy := *order
	r.records[key] = idempotencyRecord{hash: request.RequestHash, order: &copy}
	r.creates++
	return &copy, nil
}

func (r *idempotentOrderRepository) Update(context.Context, *entity.Order) error { return nil }
func (r *idempotentOrderRepository) Get(context.Context, string, int64) (*entity.Order, error) {
	return nil, nil
}
func (r *idempotentOrderRepository) List(context.Context, int64) ([]*entity.Order, error) {
	return nil, nil
}

func TestCreateOrderIsIdempotent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &idempotentOrderRepository{records: make(map[string]idempotencyRecord)}
	handler := NewOrderHandler(service.NewOrderService(repository))
	router := gin.New()
	router.Use(httpmiddleware.ErrorHandler(zap.NewNop()))
	router.POST("/api/v1/orders", handler.Create)

	first := createOrderRequest(t, router, "create-order-1", 100)
	require.Equal(t, http.StatusAccepted, first.Code)
	var firstOrder entity.Order
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstOrder))

	replay := createOrderRequest(t, router, "create-order-1", 100)
	require.Equal(t, http.StatusAccepted, replay.Code)
	var replayedOrder entity.Order
	require.NoError(t, json.Unmarshal(replay.Body.Bytes(), &replayedOrder))
	require.Equal(t, firstOrder.ID, replayedOrder.ID)
	require.Equal(t, 1, repository.creates)

	conflict := createOrderRequest(t, router, "create-order-1", 101)
	require.Equal(t, http.StatusConflict, conflict.Code)
	require.Contains(t, conflict.Body.String(), "another payload")
	require.Equal(t, 1, repository.creates)
}

func TestCreateOrderRequiresIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &idempotentOrderRepository{records: make(map[string]idempotencyRecord)}
	handler := NewOrderHandler(service.NewOrderService(repository))
	router := gin.New()
	router.Use(httpmiddleware.ErrorHandler(zap.NewNop()))
	router.POST("/api/v1/orders", handler.Create)

	response := createOrderRequest(t, router, "", 100)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "Idempotency-Key is required")
	require.Zero(t, repository.creates)
}

func createOrderRequest(t *testing.T, router http.Handler, idempotencyKey string, price int64) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(CreateOrderRequest{
		Price: price, ProductID: "product-1", CourierID: "courier-1", DeliverySlot: "slot-1",
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authUserIDHeader, "42")
	request.Header.Set(authEmailHeader, "user@example.com")
	if idempotencyKey != "" {
		request.Header.Set(idempotencyHeader, idempotencyKey)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
