package repositories

import (
	"encoding/json"
	"testing"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestSagaCommandMessages(t *testing.T) {
	order := &Order{
		ID: "order-1", UserID: 42, Email: "user@example.com", ProductID: "product-1",
		CourierID: "courier-1", DeliverySlot: "slot-1",
	}
	tests := []struct {
		name        string
		build       func(*Order, string) (*outbox.Message, error)
		topic       string
		messageType string
		assert      func(*testing.T, json.RawMessage)
	}{
		{
			name: "reserve inventory", build: newReserveInventoryMessage,
			topic: contracts.TopicWarehouseCommands, messageType: contracts.MessageReserveInventoryRequested,
			assert: func(t *testing.T, raw json.RawMessage) {
				payload := &contracts.ReserveInventory{}
				require.NoError(t, json.Unmarshal(raw, payload))
				require.Equal(t, "product-1", payload.ProductID)
				require.Equal(t, "order:order-1:inventory:reserve", payload.OperationID)
			},
		},
		{
			name: "reserve delivery", build: newReserveDeliveryMessage,
			topic: contracts.TopicDeliveryCommands, messageType: contracts.MessageReserveDeliveryRequested,
			assert: func(t *testing.T, raw json.RawMessage) {
				payload := &contracts.ReserveDelivery{}
				require.NoError(t, json.Unmarshal(raw, payload))
				require.Equal(t, "courier-1", payload.CourierID)
				require.Equal(t, "slot-1", payload.DeliverySlot)
				require.Equal(t, "order:order-1:delivery:reserve", payload.OperationID)
			},
		},
		{
			name: "release inventory", build: newReleaseInventoryMessage,
			topic: contracts.TopicWarehouseCommands, messageType: contracts.MessageReleaseInventoryRequested,
			assert: func(t *testing.T, raw json.RawMessage) {
				payload := &contracts.ReleaseInventory{}
				require.NoError(t, json.Unmarshal(raw, payload))
				require.Equal(t, "product-1", payload.ProductID)
				require.Equal(t, "order:order-1:inventory:release", payload.OperationID)
			},
		},
		{
			name: "refund payment", build: newRefundPaymentMessage,
			topic: contracts.TopicBillingCommands, messageType: contracts.MessageRefundPaymentRequested,
			assert: func(t *testing.T, raw json.RawMessage) {
				payload := &contracts.RefundPayment{}
				require.NoError(t, json.Unmarshal(raw, payload))
				require.Equal(t, "order:order-1:payment:refund", payload.OperationID)
				require.Equal(t, "order:order-1:payment", payload.OriginalOperationID)
			},
		},
		{
			name: "send notification", build: newSuccessNotificationMessage,
			topic: contracts.TopicNotificationCommands, messageType: contracts.MessageNotificationRequested,
			assert: func(t *testing.T, raw json.RawMessage) {
				payload := &contracts.SendNotification{}
				require.NoError(t, json.Unmarshal(raw, payload))
				require.Equal(t, int64(42), payload.UserID)
				require.Equal(t, "user@example.com", payload.Email)
				require.Equal(t, "completed", payload.OrderStatus)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := tt.build(order, "result-message-id")
			require.NoError(t, err)
			require.Equal(t, tt.topic, message.Topic)
			require.Equal(t, "order-1", message.MessageKey)
			require.Equal(t, tt.messageType, message.MessageType)

			envelope := &contracts.Envelope{}
			require.NoError(t, json.Unmarshal([]byte(message.Payload), envelope))
			require.Equal(t, message.ID, envelope.MessageID)
			require.Equal(t, "order-1", envelope.SagaID)
			require.Equal(t, "result-message-id", envelope.CausationID)
			tt.assert(t, envelope.Payload)
		})
	}
}

func TestFailureNotificationMessages(t *testing.T) {
	order := &Order{ID: "order-1", UserID: 42, Email: "user@example.com"}
	tests := []struct {
		stage   string
		subject string
		body    string
	}{
		{stage: "billing", subject: "Order payment failed", body: "Payment for order order-1 was declined: insufficient funds"},
		{stage: "warehouse", subject: "Order inventory reservation failed", body: "Product reservation for order order-1 failed; payment was refunded: product unavailable"},
		{stage: "delivery", subject: "Order delivery reservation failed", body: "Delivery reservation for order order-1 failed; inventory was released and payment was refunded: slot unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.stage, func(t *testing.T) {
			reason := map[string]string{
				"billing": "insufficient funds", "warehouse": "product unavailable", "delivery": "slot unavailable",
			}[tt.stage]
			message, err := newFailureNotificationMessage(order, "result-message-id", entity.FailureStage(tt.stage), reason)
			require.NoError(t, err)
			require.Equal(t, contracts.TopicNotificationCommands, message.Topic)

			envelope := &contracts.Envelope{}
			require.NoError(t, json.Unmarshal([]byte(message.Payload), envelope))
			payload := &contracts.SendNotification{}
			require.NoError(t, json.Unmarshal(envelope.Payload, payload))
			require.Equal(t, "failed", payload.OrderStatus)
			require.Equal(t, tt.subject, payload.Subject)
			require.Equal(t, tt.body, payload.Body)
		})
	}
}
