package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
)

type Processor interface {
	Reserve(context.Context, contracts.ReserveDelivery, *outbox.Message, *outbox.Message) error
}
type Handler struct{ processor Processor }

func NewHandler(processor Processor) *Handler { return &Handler{processor: processor} }
func (h *Handler) Handle(ctx context.Context, key, value []byte) error {
	envelope := &contracts.Envelope{}
	if err := json.Unmarshal(value, envelope); err != nil {
		return fmt.Errorf("decode delivery command: %w", err)
	}
	if envelope.MessageID == "" || envelope.SagaID == "" || envelope.MessageType != contracts.MessageReserveDeliveryRequested {
		return fmt.Errorf("invalid delivery envelope")
	}
	command := contracts.ReserveDelivery{}
	if err := json.Unmarshal(envelope.Payload, &command); err != nil {
		return err
	}
	if command.OrderID == "" || command.OperationID == "" || command.CourierID == "" || command.DeliverySlot == "" || string(key) != command.OrderID || envelope.SagaID != command.OrderID {
		return fmt.Errorf("invalid delivery command")
	}
	succeeded, err := resultMessage(contracts.MessageDeliveryReserved, envelope, contracts.OperationSucceeded{OrderID: command.OrderID})
	if err != nil {
		return err
	}
	failed, err := resultMessage(contracts.MessageDeliveryReservationFailed, envelope, contracts.OperationFailed{OrderID: command.OrderID, Reason: "courier slot is already reserved"})
	if err != nil {
		return err
	}
	return h.processor.Reserve(ctx, command, succeeded, failed)
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
