package entity

import "time"

type Status string
type FailureStage string

const (
	StatusPaymentPending     Status = "payment_pending"
	StatusInventoryPending   Status = "inventory_pending"
	StatusDeliveryPending    Status = "delivery_pending"
	StatusInventoryReleasing Status = "inventory_releasing"
	StatusPaymentRefunding   Status = "payment_refunding"
	StatusFailed             Status = "failed"
	StatusCompleted          Status = "completed"
)

const (
	FailureStageBilling   FailureStage = "billing"
	FailureStageWarehouse FailureStage = "warehouse"
	FailureStageDelivery  FailureStage = "delivery"
)

type Order struct {
	ID            string       `json:"id"`
	UserID        int64        `json:"userId"`
	Email         string       `json:"email"`
	Price         int64        `json:"price"`
	ProductID     string       `json:"productId"`
	CourierID     string       `json:"courierId"`
	DeliverySlot  string       `json:"deliverySlot"`
	Status        Status       `json:"status"`
	FailureStage  FailureStage `json:"failureStage,omitempty"`
	FailureReason string       `json:"failureReason,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}
