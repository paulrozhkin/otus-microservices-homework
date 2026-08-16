package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/entity"
	"github.com/stretchr/testify/require"
)

type fakeOrderRepository struct {
	order   *entity.Order
	message *outbox.Message
}

func (f *fakeOrderRepository) Create(_ context.Context, order *entity.Order, message *outbox.Message) error {
	orderCopy := *order
	messageCopy := *message
	f.order = &orderCopy
	f.message = &messageCopy
	return nil
}

func (f *fakeOrderRepository) Update(_ context.Context, order *entity.Order) error {
	copy := *order
	f.order = &copy
	return nil
}

func (f *fakeOrderRepository) Get(context.Context, string, int64) (*entity.Order, error) {
	return f.order, nil
}

func (f *fakeOrderRepository) List(context.Context, int64) ([]*entity.Order, error) {
	return []*entity.Order{f.order}, nil
}

func TestCreateOrderEnqueuesChargePayment(t *testing.T) {
	repository := &fakeOrderRepository{}
	order, err := NewOrderService(repository).Create(context.Background(), CreateOrder{
		UserID: 42, Email: "user@example.com", Price: 10000,
		ProductID: "product-1", CourierID: "courier-1", DeliverySlot: "slot-1",
	})

	require.NoError(t, err)
	require.Equal(t, entity.StatusPaymentPending, order.Status)
	require.Equal(t, "product-1", order.ProductID)
	require.Equal(t, contracts.TopicBillingCommands, repository.message.Topic)
	require.Equal(t, order.ID, repository.message.MessageKey)
	require.Equal(t, contracts.MessageChargePaymentRequested, repository.message.MessageType)

	var envelope contracts.Envelope
	require.NoError(t, json.Unmarshal([]byte(repository.message.Payload), &envelope))
	require.Equal(t, repository.message.ID, envelope.MessageID)
	require.Equal(t, order.ID, envelope.SagaID)
	require.Equal(t, order.ID, envelope.CorrelationID)

	var command contracts.ChargePayment
	require.NoError(t, json.Unmarshal(envelope.Payload, &command))
	require.Equal(t, order.ID, command.OrderID)
	require.Equal(t, "order:"+order.ID+":payment", command.OperationID)
	require.Equal(t, int64(42), command.UserID)
	require.Equal(t, int64(10000), command.Amount)
}

func TestCreateOrderValidatesSagaResources(t *testing.T) {
	tests := []struct {
		name  string
		input CreateOrder
	}{
		{name: "product", input: CreateOrder{UserID: 42, Email: "u@example.com", Price: 100, CourierID: "courier", DeliverySlot: "slot"}},
		{name: "courier", input: CreateOrder{UserID: 42, Email: "u@example.com", Price: 100, ProductID: "product", DeliverySlot: "slot"}},
		{name: "slot", input: CreateOrder{UserID: 42, Email: "u@example.com", Price: 100, ProductID: "product", CourierID: "courier"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOrderService(&fakeOrderRepository{}).Create(context.Background(), tt.input)
			require.Error(t, err)
		})
	}
}
