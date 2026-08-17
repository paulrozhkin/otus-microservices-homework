package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
)

type processorStub struct {
	paymentOperationID  string
	userID              int64
	amount              int64
	succeeded           *outbox.Message
	failed              *outbox.Message
	refundOperationID   string
	originalOperationID string
	refunded            *outbox.Message
}

func (s *processorStub) ProcessPayment(_ context.Context, operationID string, userID, amount int64, succeeded, failed *outbox.Message) error {
	s.paymentOperationID, s.userID, s.amount, s.succeeded, s.failed = operationID, userID, amount, succeeded, failed
	return nil
}

func (s *processorStub) ProcessRefund(_ context.Context, operationID, originalOperationID string, refunded *outbox.Message) error {
	s.refundOperationID, s.originalOperationID, s.refunded = operationID, originalOperationID, refunded
	return nil
}

func TestPaymentHandlerHandlesChargeCommand(t *testing.T) {
	processor := &processorStub{}
	handler := NewPaymentHandler(processor)
	command := contracts.ChargePayment{OrderID: "order-1", OperationID: "order:order-1:payment", UserID: 42, Amount: 1500}
	envelope := mustEnvelope(t, contracts.MessageChargePaymentRequested, command)

	if err := handler.Handle(context.Background(), []byte(command.OrderID), mustMarshal(t, envelope)); err != nil {
		t.Fatalf("handle charge: %v", err)
	}
	if processor.paymentOperationID != command.OperationID || processor.userID != command.UserID || processor.amount != command.Amount {
		t.Fatalf("unexpected processor arguments: %#v", processor)
	}
	assertOutboxEvent(t, processor.succeeded, contracts.MessagePaymentSucceeded, envelope.MessageID, command.OrderID)
	assertOutboxEvent(t, processor.failed, contracts.MessagePaymentFailed, envelope.MessageID, command.OrderID)
}

func TestPaymentHandlerHandlesRefundCommand(t *testing.T) {
	processor := &processorStub{}
	handler := NewPaymentHandler(processor)
	command := contracts.RefundPayment{OrderID: "order-1", OperationID: "order:order-1:refund", OriginalOperationID: "order:order-1:payment"}
	envelope := mustEnvelope(t, contracts.MessageRefundPaymentRequested, command)

	if err := handler.Handle(context.Background(), []byte(command.OrderID), mustMarshal(t, envelope)); err != nil {
		t.Fatalf("handle refund: %v", err)
	}
	if processor.refundOperationID != command.OperationID || processor.originalOperationID != command.OriginalOperationID {
		t.Fatalf("unexpected refund arguments: %#v", processor)
	}
	assertOutboxEvent(t, processor.refunded, contracts.MessagePaymentRefunded, envelope.MessageID, command.OrderID)
}

func TestPaymentHandlerRejectsMismatchedKafkaKey(t *testing.T) {
	handler := NewPaymentHandler(&processorStub{})
	envelope := mustEnvelope(t, contracts.MessageChargePaymentRequested, contracts.ChargePayment{
		OrderID: "order-1", OperationID: "payment-1", UserID: 42, Amount: 100,
	})
	if err := handler.Handle(context.Background(), []byte("another-order"), mustMarshal(t, envelope)); err == nil {
		t.Fatal("expected validation error")
	}
}

func mustEnvelope(t *testing.T, messageType string, payload any) *contracts.Envelope {
	t.Helper()
	envelope, err := contracts.NewEnvelope(messageType, "order-1", "", payload)
	if err != nil {
		t.Fatalf("new envelope: %v", err)
	}
	return envelope
}

func mustMarshal(t *testing.T, envelope *contracts.Envelope) []byte {
	t.Helper()
	encoded, err := envelope.Marshal()
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return encoded
}

func assertOutboxEvent(t *testing.T, message *outbox.Message, messageType, causationID, orderID string) {
	t.Helper()
	if message == nil || message.Topic != contracts.TopicOrderSagaEvents || message.MessageKey != orderID || message.MessageType != messageType {
		t.Fatalf("unexpected outbox message: %#v", message)
	}
	event := &contracts.Envelope{}
	if err := json.Unmarshal([]byte(message.Payload), event); err != nil {
		t.Fatalf("decode outbox event: %v", err)
	}
	if event.MessageID != message.ID || event.MessageType != messageType || event.SagaID != orderID || event.CausationID != causationID {
		t.Fatalf("unexpected event envelope: %#v", event)
	}
}
