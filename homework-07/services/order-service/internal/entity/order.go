package entity

import "time"

type Status string

const (
	StatusPending  Status = "pending"
	StatusPaid     Status = "paid"
	StatusRejected Status = "rejected"
	StatusFailed   Status = "failed"
)

type Order struct {
	ID            string    `json:"id"`
	UserID        int64     `json:"userId"`
	Email         string    `json:"email"`
	Price         int64     `json:"price"`
	Status        Status    `json:"status"`
	FailureReason string    `json:"failureReason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
