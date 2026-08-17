package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/stretchr/testify/require"
)

type processorStub struct {
	command           contracts.ReserveDelivery
	succeeded, failed *outbox.Message
}

func (s *processorStub) Reserve(_ context.Context, command contracts.ReserveDelivery, succeeded, failed *outbox.Message) error {
	s.command, s.succeeded, s.failed = command, succeeded, failed
	return nil
}
func TestHandlerRoutesDeliveryReservation(t *testing.T) {
	stub := &processorStub{}
	handler := NewHandler(stub)
	command := contracts.ReserveDelivery{OrderID: "order-1", OperationID: "delivery-1", CourierID: "courier-1", DeliverySlot: "slot-1"}
	envelope, err := contracts.NewEnvelope(contracts.MessageReserveDeliveryRequested, "order-1", "inventory-result", command)
	require.NoError(t, err)
	raw, err := envelope.Marshal()
	require.NoError(t, err)
	require.NoError(t, handler.Handle(context.Background(), []byte("order-1"), raw))
	require.Equal(t, command, stub.command)
	assertResult(t, stub.succeeded, contracts.MessageDeliveryReserved, envelope.MessageID)
	assertResult(t, stub.failed, contracts.MessageDeliveryReservationFailed, envelope.MessageID)
}
func TestHandlerRejectsMismatchedKey(t *testing.T) {
	handler := NewHandler(&processorStub{})
	envelope, err := contracts.NewEnvelope(contracts.MessageReserveDeliveryRequested, "order-1", "", contracts.ReserveDelivery{OrderID: "order-1", OperationID: "delivery-1", CourierID: "courier-1", DeliverySlot: "slot-1"})
	require.NoError(t, err)
	raw, err := envelope.Marshal()
	require.NoError(t, err)
	require.Error(t, handler.Handle(context.Background(), []byte("another-order"), raw))
}
func assertResult(t *testing.T, message *outbox.Message, messageType, causationID string) {
	t.Helper()
	require.Equal(t, contracts.TopicOrderSagaEvents, message.Topic)
	event := contracts.Envelope{}
	require.NoError(t, json.Unmarshal([]byte(message.Payload), &event))
	require.Equal(t, messageType, event.MessageType)
	require.Equal(t, causationID, event.CausationID)
}
