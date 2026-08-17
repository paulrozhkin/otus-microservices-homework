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
	reserve                     contracts.ReserveInventory
	release                     contracts.ReleaseInventory
	succeeded, failed, released *outbox.Message
}

func (s *processorStub) Reserve(_ context.Context, c contracts.ReserveInventory, ok, failed *outbox.Message) error {
	s.reserve, s.succeeded, s.failed = c, ok, failed
	return nil
}
func (s *processorStub) Release(_ context.Context, c contracts.ReleaseInventory, released *outbox.Message) error {
	s.release, s.released = c, released
	return nil
}
func TestHandlerRoutesReserveAndRelease(t *testing.T) {
	tests := []struct {
		name, kind string
		payload    any
	}{
		{"reserve", contracts.MessageReserveInventoryRequested, contracts.ReserveInventory{OrderID: "order-1", OperationID: "reserve-1", ProductID: "product-1"}},
		{"release", contracts.MessageReleaseInventoryRequested, contracts.ReleaseInventory{OrderID: "order-1", OperationID: "release-1", ProductID: "product-1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &processorStub{}
			handler := NewHandler(stub)
			envelope, err := contracts.NewEnvelope(tt.kind, "order-1", "previous", tt.payload)
			require.NoError(t, err)
			raw, err := envelope.Marshal()
			require.NoError(t, err)
			require.NoError(t, handler.Handle(context.Background(), []byte("order-1"), raw))
			if tt.name == "reserve" {
				require.Equal(t, "reserve-1", stub.reserve.OperationID)
				assertEvent(t, stub.succeeded, contracts.MessageInventoryReserved, envelope.MessageID)
				assertEvent(t, stub.failed, contracts.MessageInventoryReservationFailed, envelope.MessageID)
			} else {
				require.Equal(t, "release-1", stub.release.OperationID)
				assertEvent(t, stub.released, contracts.MessageInventoryReleased, envelope.MessageID)
			}
		})
	}
}
func assertEvent(t *testing.T, message *outbox.Message, messageType, causationID string) {
	t.Helper()
	require.Equal(t, contracts.TopicOrderSagaEvents, message.Topic)
	event := contracts.Envelope{}
	require.NoError(t, json.Unmarshal([]byte(message.Payload), &event))
	require.Equal(t, messageType, event.MessageType)
	require.Equal(t, causationID, event.CausationID)
}
