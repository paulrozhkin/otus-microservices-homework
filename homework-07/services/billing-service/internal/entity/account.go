package entity

import "time"

type Account struct {
	UserID    int64     `json:"userId"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Operation struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"userId"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

const (
	OperationDeposit    = "deposit"
	OperationWithdrawal = "withdrawal"
	OperationPayment    = "payment"
	OperationRefund     = "refund"
)
