package entity

import "time"

type Status string

const (
	StatusPaymentPending Status = "payment_pending"
	StatusFailed         Status = "failed"
	StatusCompleted      Status = "completed"
)

type Order struct {
	ID            string    `json:"id"`
	UserID        int64     `json:"userId"`
	Email         string    `json:"email"`
	Price         int64     `json:"price"`
	ProductID     string    `json:"productId"`
	CourierID     string    `json:"courierId"`
	DeliverySlot  string    `json:"deliverySlot"`
	Status        Status    `json:"status"`
	FailureReason string    `json:"failureReason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
