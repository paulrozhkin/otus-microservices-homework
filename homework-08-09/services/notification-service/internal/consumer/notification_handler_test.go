package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/entity"
	"github.com/stretchr/testify/require"
)

type fakeNotificationRepository struct {
	notification *entity.Notification
	createCalls  int
}

func (f *fakeNotificationRepository) Create(_ context.Context, notification *entity.Notification) error {
	f.createCalls++
	copy := *notification
	f.notification = &copy
	return nil
}

func (f *fakeNotificationRepository) List(context.Context, int64) ([]*entity.Notification, error) {
	return nil, nil
}

func TestHandleNotificationRequested(t *testing.T) {
	repository := &fakeNotificationRepository{}
	handler := NewNotificationHandler(repository)
	occurredAt := time.Now().UTC().Truncate(time.Millisecond)
	payload, err := json.Marshal(NotificationRequested{
		EventID: "event-1", EventType: NotificationRequestedV1, OrderID: "order-1",
		UserID: 42, Email: "user@example.com", OrderStatus: "paid",
		Subject: "Order paid", Body: "Success", OccurredAt: occurredAt,
	})
	require.NoError(t, err)

	require.NoError(t, handler.Handle(context.Background(), []byte("order-1"), payload))
	require.Equal(t, 1, repository.createCalls)
	require.Equal(t, "event-1", repository.notification.EventID)
	require.Equal(t, int64(42), repository.notification.UserID)
	require.Equal(t, "paid", repository.notification.OrderStatus)
	require.Equal(t, occurredAt, repository.notification.CreatedAt)
}

func TestHandleRejectsInvalidEvent(t *testing.T) {
	repository := &fakeNotificationRepository{}
	handler := NewNotificationHandler(repository)

	err := handler.Handle(context.Background(), nil, []byte(`{"eventType":"unknown"}`))
	require.Error(t, err)
	require.Zero(t, repository.createCalls)
}
