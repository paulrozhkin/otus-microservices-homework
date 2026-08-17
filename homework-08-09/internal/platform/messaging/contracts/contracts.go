package contracts

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	TopicBillingCommands      = "billing.commands.v1"
	TopicWarehouseCommands    = "warehouse.commands.v1"
	TopicDeliveryCommands     = "delivery.commands.v1"
	TopicOrderSagaEvents      = "order.saga.events.v1"
	TopicNotificationCommands = "notification.commands.v1"
)

const (
	MessageChargePaymentRequested = "billing.payment.charge.requested.v1"
	MessageRefundPaymentRequested = "billing.payment.refund.requested.v1"
	MessagePaymentSucceeded       = "billing.payment.succeeded.v1"
	MessagePaymentFailed          = "billing.payment.failed.v1"
	MessagePaymentRefunded        = "billing.payment.refunded.v1"

	MessageReserveInventoryRequested  = "warehouse.inventory.reserve.requested.v1"
	MessageReleaseInventoryRequested  = "warehouse.inventory.release.requested.v1"
	MessageInventoryReserved          = "warehouse.inventory.reserved.v1"
	MessageInventoryReservationFailed = "warehouse.inventory.reservation-failed.v1"
	MessageInventoryReleased          = "warehouse.inventory.released.v1"

	MessageReserveDeliveryRequested  = "delivery.slot.reserve.requested.v1"
	MessageDeliveryReserved          = "delivery.slot.reserved.v1"
	MessageDeliveryReservationFailed = "delivery.slot.reservation-failed.v1"

	MessageNotificationRequested = "notification.requested.v1"
)

type Envelope struct {
	MessageID     string          `json:"messageId"`
	MessageType   string          `json:"messageType"`
	SagaID        string          `json:"sagaId"`
	CorrelationID string          `json:"correlationId"`
	CausationID   string          `json:"causationId,omitempty"`
	OccurredAt    time.Time       `json:"occurredAt"`
	Payload       json.RawMessage `json:"payload"`
}

func NewEnvelope(messageType, sagaID, causationID string, payload any) (*Envelope, error) {
	if messageType == "" {
		return nil, fmt.Errorf("message type is required")
	}
	if sagaID == "" {
		return nil, fmt.Errorf("saga id is required")
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal message payload: %w", err)
	}

	return &Envelope{
		MessageID:     uuid.NewString(),
		MessageType:   messageType,
		SagaID:        sagaID,
		CorrelationID: sagaID,
		CausationID:   causationID,
		OccurredAt:    time.Now().UTC(),
		Payload:       rawPayload,
	}, nil
}

func (e *Envelope) Marshal() ([]byte, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("marshal message envelope: %w", err)
	}
	return payload, nil
}

type ChargePayment struct {
	OrderID     string `json:"orderId"`
	OperationID string `json:"operationId"`
	UserID      int64  `json:"userId"`
	Amount      int64  `json:"amount"`
}

type RefundPayment struct {
	OrderID             string `json:"orderId"`
	OperationID         string `json:"operationId"`
	OriginalOperationID string `json:"originalOperationId"`
}

type ReserveInventory struct {
	OrderID     string `json:"orderId"`
	OperationID string `json:"operationId"`
	ProductID   string `json:"productId"`
}

type ReleaseInventory struct {
	OrderID     string `json:"orderId"`
	OperationID string `json:"operationId"`
	ProductID   string `json:"productId"`
}

type ReserveDelivery struct {
	OrderID      string `json:"orderId"`
	OperationID  string `json:"operationId"`
	CourierID    string `json:"courierId"`
	DeliverySlot string `json:"deliverySlot"`
}

type OperationSucceeded struct {
	OrderID string `json:"orderId"`
}

type OperationFailed struct {
	OrderID string `json:"orderId"`
	Reason  string `json:"reason"`
}

type SendNotification struct {
	OrderID     string `json:"orderId"`
	UserID      int64  `json:"userId"`
	Email       string `json:"email"`
	OrderStatus string `json:"orderStatus"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
}
