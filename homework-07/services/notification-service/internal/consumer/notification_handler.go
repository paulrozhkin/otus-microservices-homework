package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/notification-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/notification-service/internal/repositories"
)

const NotificationRequestedV1 = "notification.requested.v1"

type NotificationRequested struct {
	EventID     string    `json:"eventId"`
	EventType   string    `json:"eventType"`
	OrderID     string    `json:"orderId"`
	UserID      int64     `json:"userId"`
	Email       string    `json:"email"`
	OrderStatus string    `json:"orderStatus"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	OccurredAt  time.Time `json:"occurredAt"`
}

type NotificationHandler struct {
	repository repositories.NotificationRepository
}

func NewNotificationHandler(repository repositories.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repository: repository}
}

func (h *NotificationHandler) Handle(ctx context.Context, _, value []byte) error {
	event := &NotificationRequested{}
	if err := json.Unmarshal(value, event); err != nil {
		return fmt.Errorf("decode notification requested event: %w", err)
	}
	if err := validateEvent(event); err != nil {
		return err
	}
	return h.repository.Create(ctx, &entity.Notification{
		EventID: event.EventID, OrderID: event.OrderID, UserID: event.UserID,
		Email: event.Email, OrderStatus: event.OrderStatus, Subject: event.Subject,
		Body: event.Body, CreatedAt: event.OccurredAt,
	})
}

func validateEvent(event *NotificationRequested) error {
	if event.EventType != NotificationRequestedV1 {
		return fmt.Errorf("unsupported event type %q", event.EventType)
	}
	if event.EventID == "" || event.OrderID == "" || event.UserID <= 0 || event.Email == "" || event.Subject == "" || event.Body == "" || event.OccurredAt.IsZero() {
		return fmt.Errorf("invalid %s event", NotificationRequestedV1)
	}
	return nil
}
