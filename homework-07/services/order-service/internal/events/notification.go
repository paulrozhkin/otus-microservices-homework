package events

import "time"

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
