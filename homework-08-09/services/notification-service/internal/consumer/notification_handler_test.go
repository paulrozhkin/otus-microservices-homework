package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
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
	payload := notificationEnvelope(t, "order-1", occurredAt)

	require.NoError(t, handler.Handle(context.Background(), []byte("order-1"), payload))
	require.Equal(t, 1, repository.createCalls)
	require.Equal(t, "message-1", repository.notification.EventID)
	require.Equal(t, "order-1", repository.notification.OrderID)
	require.Equal(t, int64(42), repository.notification.UserID)
	require.Equal(t, "completed", repository.notification.OrderStatus)
	require.Equal(t, occurredAt, repository.notification.CreatedAt)
}

func TestHandleRejectsUnsupportedMessage(t *testing.T) {
	repository := &fakeNotificationRepository{}
	handler := NewNotificationHandler(repository)
	payload := notificationEnvelope(t, "order-1", time.Now().UTC())
	var envelope contracts.Envelope
	require.NoError(t, json.Unmarshal(payload, &envelope))
	envelope.MessageType = "unknown"
	payload, err := json.Marshal(envelope)
	require.NoError(t, err)

	err = handler.Handle(context.Background(), []byte("order-1"), payload)
	require.Error(t, err)
	require.Zero(t, repository.createCalls)
}

func TestHandleRejectsMismatchedKafkaKey(t *testing.T) {
	repository := &fakeNotificationRepository{}
	handler := NewNotificationHandler(repository)

	err := handler.Handle(context.Background(), []byte("another-order"), notificationEnvelope(t, "order-1", time.Now().UTC()))
	require.Error(t, err)
	require.Zero(t, repository.createCalls)
}

func TestHandleRejectsInvalidPayload(t *testing.T) {
	repository := &fakeNotificationRepository{}
	handler := NewNotificationHandler(repository)
	payload := notificationEnvelope(t, "order-1", time.Now().UTC())
	var envelope contracts.Envelope
	require.NoError(t, json.Unmarshal(payload, &envelope))
	envelope.Payload = json.RawMessage(`{"orderId":"order-1"}`)
	payload, err := json.Marshal(envelope)
	require.NoError(t, err)

	err = handler.Handle(context.Background(), []byte("order-1"), payload)
	require.Error(t, err)
	require.Zero(t, repository.createCalls)
}

func notificationEnvelope(t *testing.T, orderID string, occurredAt time.Time) []byte {
	t.Helper()
	payload, err := json.Marshal(contracts.SendNotification{
		OrderID: orderID, UserID: 42, Email: "user@example.com", OrderStatus: "completed",
		Subject: "Order completed", Body: "Success",
	})
	require.NoError(t, err)

	envelope := contracts.Envelope{
		MessageID: "message-1", MessageType: contracts.MessageNotificationRequested,
		SagaID: orderID, CorrelationID: orderID, CausationID: "delivery-event-1",
		OccurredAt: occurredAt, Payload: payload,
	}
	encoded, err := envelope.Marshal()
	require.NoError(t, err)
	return encoded
}
