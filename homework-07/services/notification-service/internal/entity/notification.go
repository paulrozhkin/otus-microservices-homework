package entity

import "time"

type Notification struct {
	EventID     string    `json:"eventId"`
	OrderID     string    `json:"orderId"`
	UserID      int64     `json:"userId"`
	Email       string    `json:"email"`
	OrderStatus string    `json:"orderStatus"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"createdAt"`
}
