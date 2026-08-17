package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
)

type PaymentProcessor interface {
	ProcessPayment(context.Context, string, int64, int64, *outbox.Message, *outbox.Message) error
	ProcessRefund(context.Context, string, string, *outbox.Message) error
}

type PaymentHandler struct {
	processor PaymentProcessor
}

func NewPaymentHandler(processor PaymentProcessor) *PaymentHandler {
	return &PaymentHandler{processor: processor}
}

func (h *PaymentHandler) Handle(ctx context.Context, key, value []byte) error {
	envelope := &contracts.Envelope{}
	if err := json.Unmarshal(value, envelope); err != nil {
		return fmt.Errorf("decode billing command envelope: %w", err)
	}
	if envelope.MessageID == "" || envelope.SagaID == "" {
		return fmt.Errorf("invalid billing command envelope")
	}

	switch envelope.MessageType {
	case contracts.MessageChargePaymentRequested:
		return h.handleCharge(ctx, string(key), envelope)
	case contracts.MessageRefundPaymentRequested:
		return h.handleRefund(ctx, string(key), envelope)
	default:
		return fmt.Errorf("unsupported billing message type %q", envelope.MessageType)
	}
}

func (h *PaymentHandler) handleCharge(ctx context.Context, key string, envelope *contracts.Envelope) error {
	command := &contracts.ChargePayment{}
	if err := json.Unmarshal(envelope.Payload, command); err != nil {
		return fmt.Errorf("decode charge payment payload: %w", err)
	}
	if command.OrderID == "" || command.OperationID == "" || command.UserID <= 0 || command.Amount <= 0 || key != command.OrderID || envelope.SagaID != command.OrderID {
		return fmt.Errorf("invalid charge payment command")
	}

	succeeded, err := newOutboxMessage(contracts.MessagePaymentSucceeded, envelope, command.OrderID, contracts.OperationSucceeded{OrderID: command.OrderID})
	if err != nil {
		return err
	}
	failed, err := newOutboxMessage(contracts.MessagePaymentFailed, envelope, command.OrderID, contracts.OperationFailed{
		OrderID: command.OrderID,
		Reason:  "account not found or insufficient funds",
	})
	if err != nil {
		return err
	}
	return h.processor.ProcessPayment(ctx, command.OperationID, command.UserID, command.Amount, succeeded, failed)
}

func (h *PaymentHandler) handleRefund(ctx context.Context, key string, envelope *contracts.Envelope) error {
	command := &contracts.RefundPayment{}
	if err := json.Unmarshal(envelope.Payload, command); err != nil {
		return fmt.Errorf("decode refund payment payload: %w", err)
	}
	if command.OrderID == "" || command.OperationID == "" || command.OriginalOperationID == "" || key != command.OrderID || envelope.SagaID != command.OrderID {
		return fmt.Errorf("invalid refund payment command")
	}

	refunded, err := newOutboxMessage(contracts.MessagePaymentRefunded, envelope, command.OrderID, contracts.OperationSucceeded{OrderID: command.OrderID})
	if err != nil {
		return err
	}
	return h.processor.ProcessRefund(ctx, command.OperationID, command.OriginalOperationID, refunded)
}

func newOutboxMessage(messageType string, command *contracts.Envelope, orderID string, payload any) (*outbox.Message, error) {
	event, err := contracts.NewEnvelope(messageType, command.SagaID, command.MessageID, payload)
	if err != nil {
		return nil, err
	}
	encoded, err := event.Marshal()
	if err != nil {
		return nil, err
	}
	return &outbox.Message{
		ID:          event.MessageID,
		Topic:       contracts.TopicOrderSagaEvents,
		MessageKey:  orderID,
		MessageType: messageType,
		Payload:     string(encoded),
	}, nil
}
