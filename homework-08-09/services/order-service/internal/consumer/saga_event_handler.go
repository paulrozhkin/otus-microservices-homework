package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
)

type SagaResultProcessor interface {
	ApplyPaymentSucceeded(context.Context, string, string) error
	ApplyPaymentFailed(context.Context, string, string) error
	ApplyInventoryReserved(context.Context, string, string) error
	ApplyInventoryReservationFailed(context.Context, string, string, string) error
	ApplyDeliveryReserved(context.Context, string, string) error
	ApplyDeliveryReservationFailed(context.Context, string, string, string) error
	ApplyInventoryReleased(context.Context, string, string) error
	ApplyPaymentRefunded(context.Context, string) error
}

type SagaEventHandler struct {
	processor SagaResultProcessor
}

func NewSagaEventHandler(processor SagaResultProcessor) *SagaEventHandler {
	return &SagaEventHandler{processor: processor}
}

func (h *SagaEventHandler) Handle(ctx context.Context, key, value []byte) error {
	envelope := &contracts.Envelope{}
	if err := json.Unmarshal(value, envelope); err != nil {
		return fmt.Errorf("decode saga event envelope: %w", err)
	}
	if envelope.MessageID == "" || envelope.SagaID == "" {
		return fmt.Errorf("invalid saga event envelope")
	}

	switch envelope.MessageType {
	case contracts.MessagePaymentSucceeded:
		orderID, err := decodeSucceededResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyPaymentSucceeded(ctx, orderID, envelope.MessageID)

	case contracts.MessagePaymentFailed:
		orderID, reason, err := decodeFailedResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyPaymentFailed(ctx, orderID, reason)

	case contracts.MessageInventoryReserved:
		orderID, err := decodeSucceededResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyInventoryReserved(ctx, orderID, envelope.MessageID)

	case contracts.MessageInventoryReservationFailed:
		orderID, reason, err := decodeFailedResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyInventoryReservationFailed(ctx, orderID, reason, envelope.MessageID)

	case contracts.MessageDeliveryReserved:
		orderID, err := decodeSucceededResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyDeliveryReserved(ctx, orderID, envelope.MessageID)

	case contracts.MessageDeliveryReservationFailed:
		orderID, reason, err := decodeFailedResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyDeliveryReservationFailed(ctx, orderID, reason, envelope.MessageID)

	case contracts.MessageInventoryReleased:
		orderID, err := decodeSucceededResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyInventoryReleased(ctx, orderID, envelope.MessageID)

	case contracts.MessagePaymentRefunded:
		orderID, err := decodeSucceededResult(string(key), envelope)
		if err != nil {
			return err
		}
		return h.processor.ApplyPaymentRefunded(ctx, orderID)

	default:
		return fmt.Errorf("unsupported saga event type %q", envelope.MessageType)
	}
}

func decodeSucceededResult(key string, envelope *contracts.Envelope) (string, error) {
	result := &contracts.OperationSucceeded{}
	if err := json.Unmarshal(envelope.Payload, result); err != nil {
		return "", fmt.Errorf("decode successful saga result: %w", err)
	}
	if err := validateResult(key, envelope.SagaID, result.OrderID); err != nil {
		return "", err
	}
	return result.OrderID, nil
}

func decodeFailedResult(key string, envelope *contracts.Envelope) (string, string, error) {
	result := &contracts.OperationFailed{}
	if err := json.Unmarshal(envelope.Payload, result); err != nil {
		return "", "", fmt.Errorf("decode failed saga result: %w", err)
	}
	if err := validateResult(key, envelope.SagaID, result.OrderID); err != nil {
		return "", "", err
	}
	if result.Reason == "" {
		return "", "", fmt.Errorf("saga failure reason is required")
	}
	return result.OrderID, result.Reason, nil
}

func validateResult(key, sagaID, orderID string) error {
	if orderID == "" || key != orderID || sagaID != orderID {
		return fmt.Errorf("invalid saga result")
	}
	return nil
}
