package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/entity"
	businessmetrics "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/metrics"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/repositories"
)

type NotificationHandler struct {
	repository repositories.NotificationRepository
}

func NewNotificationHandler(repository repositories.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repository: repository}
}

func (h *NotificationHandler) Handle(ctx context.Context, key, value []byte) error {
	envelope := &contracts.Envelope{}
	if err := json.Unmarshal(value, envelope); err != nil {
		return fmt.Errorf("decode notification command: %w", err)
	}
	if err := validateEnvelope(string(key), envelope); err != nil {
		return err
	}

	command := &contracts.SendNotification{}
	if err := json.Unmarshal(envelope.Payload, command); err != nil {
		return fmt.Errorf("decode notification payload: %w", err)
	}
	if err := validateCommand(string(key), envelope.SagaID, command); err != nil {
		return err
	}

	if err := h.repository.Create(ctx, &entity.Notification{
		EventID: envelope.MessageID, OrderID: command.OrderID, UserID: command.UserID,
		Email: command.Email, OrderStatus: command.OrderStatus, Subject: command.Subject,
		Body: command.Body, CreatedAt: envelope.OccurredAt,
	}); err != nil {
		return err
	}
	businessmetrics.NotificationProcessed(command.OrderStatus)
	return nil
}

func validateEnvelope(key string, envelope *contracts.Envelope) error {
	if envelope.MessageID == "" || envelope.MessageType != contracts.MessageNotificationRequested ||
		envelope.SagaID == "" || envelope.CorrelationID == "" || envelope.OccurredAt.IsZero() ||
		len(envelope.Payload) == 0 || key != envelope.SagaID {
		return fmt.Errorf("invalid notification envelope")
	}
	return nil
}

func validateCommand(key, sagaID string, command *contracts.SendNotification) error {
	if command.OrderID == "" || command.UserID <= 0 || command.Email == "" ||
		command.OrderStatus == "" || command.Subject == "" || command.Body == "" ||
		key != command.OrderID || sagaID != command.OrderID {
		return fmt.Errorf("invalid notification command")
	}
	return nil
}
