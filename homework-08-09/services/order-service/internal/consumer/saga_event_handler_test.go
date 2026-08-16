package consumer

import (
	"context"
	"testing"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/stretchr/testify/require"
)

type sagaResultProcessorStub struct {
	call        string
	orderID     string
	reason      string
	causationID string
}

func (s *sagaResultProcessorStub) record(call, orderID, reason, causationID string) error {
	s.call, s.orderID, s.reason, s.causationID = call, orderID, reason, causationID
	return nil
}

func (s *sagaResultProcessorStub) ApplyPaymentSucceeded(_ context.Context, orderID, causationID string) error {
	return s.record("payment_succeeded", orderID, "", causationID)
}
func (s *sagaResultProcessorStub) ApplyPaymentFailed(_ context.Context, orderID, reason, causationID string) error {
	return s.record("payment_failed", orderID, reason, causationID)
}
func (s *sagaResultProcessorStub) ApplyInventoryReserved(_ context.Context, orderID, causationID string) error {
	return s.record("inventory_reserved", orderID, "", causationID)
}
func (s *sagaResultProcessorStub) ApplyInventoryReservationFailed(_ context.Context, orderID, reason, causationID string) error {
	return s.record("inventory_failed", orderID, reason, causationID)
}
func (s *sagaResultProcessorStub) ApplyDeliveryReserved(_ context.Context, orderID, causationID string) error {
	return s.record("delivery_reserved", orderID, "", causationID)
}
func (s *sagaResultProcessorStub) ApplyDeliveryReservationFailed(_ context.Context, orderID, reason, causationID string) error {
	return s.record("delivery_failed", orderID, reason, causationID)
}
func (s *sagaResultProcessorStub) ApplyInventoryReleased(_ context.Context, orderID, causationID string) error {
	return s.record("inventory_released", orderID, "", causationID)
}
func (s *sagaResultProcessorStub) ApplyPaymentRefunded(_ context.Context, orderID, causationID string) error {
	return s.record("payment_refunded", orderID, "", causationID)
}

func TestSagaEventHandlerRoutesResults(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
		payload     any
		call        string
		reason      string
		caused      bool
	}{
		{name: "payment succeeded", messageType: contracts.MessagePaymentSucceeded, payload: contracts.OperationSucceeded{OrderID: "order-1"}, call: "payment_succeeded", caused: true},
		{name: "payment failed", messageType: contracts.MessagePaymentFailed, payload: contracts.OperationFailed{OrderID: "order-1", Reason: "insufficient funds"}, call: "payment_failed", reason: "insufficient funds", caused: true},
		{name: "inventory reserved", messageType: contracts.MessageInventoryReserved, payload: contracts.OperationSucceeded{OrderID: "order-1"}, call: "inventory_reserved", caused: true},
		{name: "inventory failed", messageType: contracts.MessageInventoryReservationFailed, payload: contracts.OperationFailed{OrderID: "order-1", Reason: "product unavailable"}, call: "inventory_failed", reason: "product unavailable", caused: true},
		{name: "delivery reserved", messageType: contracts.MessageDeliveryReserved, payload: contracts.OperationSucceeded{OrderID: "order-1"}, call: "delivery_reserved", caused: true},
		{name: "delivery failed", messageType: contracts.MessageDeliveryReservationFailed, payload: contracts.OperationFailed{OrderID: "order-1", Reason: "slot unavailable"}, call: "delivery_failed", reason: "slot unavailable", caused: true},
		{name: "inventory released", messageType: contracts.MessageInventoryReleased, payload: contracts.OperationSucceeded{OrderID: "order-1"}, call: "inventory_released", caused: true},
		{name: "payment refunded", messageType: contracts.MessagePaymentRefunded, payload: contracts.OperationSucceeded{OrderID: "order-1"}, call: "payment_refunded", caused: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &sagaResultProcessorStub{}
			handler := NewSagaEventHandler(processor)
			envelope := newResultEnvelope(t, tt.messageType, tt.payload)

			require.NoError(t, handler.Handle(context.Background(), []byte("order-1"), marshalEnvelope(t, envelope)))
			require.Equal(t, tt.call, processor.call)
			require.Equal(t, "order-1", processor.orderID)
			require.Equal(t, tt.reason, processor.reason)
			if tt.caused {
				require.Equal(t, envelope.MessageID, processor.causationID)
			} else {
				require.Empty(t, processor.causationID)
			}
		})
	}
}

func TestSagaEventHandlerRejectsMismatchedKafkaKey(t *testing.T) {
	handler := NewSagaEventHandler(&sagaResultProcessorStub{})
	envelope := newResultEnvelope(t, contracts.MessagePaymentSucceeded, contracts.OperationSucceeded{OrderID: "order-1"})
	require.Error(t, handler.Handle(context.Background(), []byte("another-order"), marshalEnvelope(t, envelope)))
}

func TestSagaEventHandlerRejectsFailureWithoutReason(t *testing.T) {
	handler := NewSagaEventHandler(&sagaResultProcessorStub{})
	envelope := newResultEnvelope(t, contracts.MessageInventoryReservationFailed, contracts.OperationFailed{OrderID: "order-1"})
	require.Error(t, handler.Handle(context.Background(), []byte("order-1"), marshalEnvelope(t, envelope)))
}

func newResultEnvelope(t *testing.T, messageType string, payload any) *contracts.Envelope {
	t.Helper()
	envelope, err := contracts.NewEnvelope(messageType, "order-1", "command-id", payload)
	require.NoError(t, err)
	return envelope
}

func marshalEnvelope(t *testing.T, envelope *contracts.Envelope) []byte {
	t.Helper()
	payload, err := envelope.Marshal()
	require.NoError(t, err)
	return payload
}
