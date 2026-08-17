package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
)

type Processor interface {
	Reserve(context.Context, contracts.ReserveInventory, *outbox.Message, *outbox.Message) error
	Release(context.Context, contracts.ReleaseInventory, *outbox.Message) error
}
type Handler struct{ processor Processor }

func NewHandler(processor Processor) *Handler { return &Handler{processor: processor} }
func (h *Handler) Handle(ctx context.Context, key, value []byte) error {
	envelope := &contracts.Envelope{}
	if err := json.Unmarshal(value, envelope); err != nil {
		return fmt.Errorf("decode warehouse command: %w", err)
	}
	if envelope.MessageID == "" || envelope.SagaID == "" {
		return fmt.Errorf("invalid warehouse envelope")
	}
	switch envelope.MessageType {
	case contracts.MessageReserveInventoryRequested:
		command := contracts.ReserveInventory{}
		if err := json.Unmarshal(envelope.Payload, &command); err != nil {
			return err
		}
		if err := validate(string(key), envelope.SagaID, command.OrderID, command.OperationID, command.ProductID); err != nil {
			return err
		}
		success, err := resultMessage(contracts.MessageInventoryReserved, envelope, contracts.OperationSucceeded{OrderID: command.OrderID})
		if err != nil {
			return err
		}
		failed, err := resultMessage(contracts.MessageInventoryReservationFailed, envelope, contracts.OperationFailed{OrderID: command.OrderID, Reason: "product is already reserved"})
		if err != nil {
			return err
		}
		return h.processor.Reserve(ctx, command, success, failed)
	case contracts.MessageReleaseInventoryRequested:
		command := contracts.ReleaseInventory{}
		if err := json.Unmarshal(envelope.Payload, &command); err != nil {
			return err
		}
		if err := validate(string(key), envelope.SagaID, command.OrderID, command.OperationID, command.ProductID); err != nil {
			return err
		}
		released, err := resultMessage(contracts.MessageInventoryReleased, envelope, contracts.OperationSucceeded{OrderID: command.OrderID})
		if err != nil {
			return err
		}
		return h.processor.Release(ctx, command, released)
	default:
		return fmt.Errorf("unsupported warehouse command %q", envelope.MessageType)
	}
}
func validate(key, sagaID, orderID, operationID, productID string) error {
	if orderID == "" || operationID == "" || productID == "" || key != orderID || sagaID != orderID {
		return fmt.Errorf("invalid warehouse command")
	}
	return nil
}
func resultMessage(messageType string, command *contracts.Envelope, payload any) (*outbox.Message, error) {
	event, err := contracts.NewEnvelope(messageType, command.SagaID, command.MessageID, payload)
	if err != nil {
		return nil, err
	}
	encoded, err := event.Marshal()
	if err != nil {
		return nil, err
	}
	return &outbox.Message{ID: event.MessageID, Topic: contracts.TopicOrderSagaEvents, MessageKey: command.SagaID, MessageType: messageType, Payload: string(encoded)}, nil
}
