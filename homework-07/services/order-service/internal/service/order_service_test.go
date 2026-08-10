package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/order-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/order-service/internal/events"
	"github.com/stretchr/testify/require"
)

type fakeOrderRepository struct{ order *entity.Order }

func (f *fakeOrderRepository) Create(_ context.Context, order *entity.Order) error {
	copy := *order
	f.order = &copy
	return nil
}
func (f *fakeOrderRepository) Update(_ context.Context, order *entity.Order) error {
	copy := *order
	f.order = &copy
	return nil
}
func (f *fakeOrderRepository) Get(context.Context, string, int64) (*entity.Order, error) {
	return f.order, nil
}
func (f *fakeOrderRepository) List(context.Context, int64) ([]*entity.Order, error) {
	return []*entity.Order{f.order}, nil
}

type fakeBilling struct {
	err     error
	orderID string
}

func (f *fakeBilling) Pay(_ context.Context, orderID string, _, _ int64) error {
	f.orderID = orderID
	return f.err
}

type fakePublisher struct{ key, value []byte }

func (f *fakePublisher) Publish(_ context.Context, key, value []byte) error {
	f.key = append([]byte(nil), key...)
	f.value = append([]byte(nil), value...)
	return nil
}

func TestCreatePaidOrderPublishesSuccessNotification(t *testing.T) {
	repository, billing, publisher := &fakeOrderRepository{}, &fakeBilling{}, &fakePublisher{}
	order, err := NewOrderService(repository, billing, publisher).Create(context.Background(), 42, "user@example.com", 10000)
	require.NoError(t, err)
	require.Equal(t, entity.StatusPaid, order.Status)
	require.Equal(t, order.ID, billing.orderID)
	require.Equal(t, order.ID, string(publisher.key))
	var event events.NotificationRequested
	require.NoError(t, json.Unmarshal(publisher.value, &event))
	require.Equal(t, events.NotificationRequestedV1, event.EventType)
	require.Equal(t, string(entity.StatusPaid), event.OrderStatus)
	require.Equal(t, "user@example.com", event.Email)
}

func TestCreateRejectedOrderPublishesFailureNotification(t *testing.T) {
	repository := &fakeOrderRepository{}
	publisher := &fakePublisher{}
	order, err := NewOrderService(repository, &fakeBilling{err: apperror.ErrInsufficientFunds}, publisher).Create(context.Background(), 42, "user@example.com", 10000)
	require.NoError(t, err)
	require.Equal(t, entity.StatusRejected, order.Status)
	require.Equal(t, apperror.ErrInsufficientFunds.Error(), order.FailureReason)
	var event events.NotificationRequested
	require.NoError(t, json.Unmarshal(publisher.value, &event))
	require.Equal(t, string(entity.StatusRejected), event.OrderStatus)
	require.Contains(t, event.Body, "insufficient funds")
}
