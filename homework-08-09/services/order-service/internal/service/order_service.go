package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/repositories"
)

type CreateOrder struct {
	IdempotencyKey string
	UserID         int64
	Email          string
	Price          int64
	ProductID      string
	CourierID      string
	DeliverySlot   string
}

type OrderService struct {
	repository repositories.OrderRepository
}

func NewOrderService(repository repositories.OrderRepository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input CreateOrder) (*entity.Order, error) {
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.IdempotencyKey == "" {
		return nil, fmt.Errorf("%w: Idempotency-Key is required", apperror.ErrInvalidOperation)
	}
	if len(input.IdempotencyKey) > 255 {
		return nil, fmt.Errorf("%w: Idempotency-Key must not exceed 255 characters", apperror.ErrInvalidOperation)
	}
	if input.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be positive", apperror.ErrInvalidOperation)
	}
	if strings.TrimSpace(input.ProductID) == "" {
		return nil, fmt.Errorf("%w: productId is required", apperror.ErrInvalidOperation)
	}
	if strings.TrimSpace(input.CourierID) == "" {
		return nil, fmt.Errorf("%w: courierId is required", apperror.ErrInvalidOperation)
	}
	if strings.TrimSpace(input.DeliverySlot) == "" {
		return nil, fmt.Errorf("%w: deliverySlot is required", apperror.ErrInvalidOperation)
	}

	now := time.Now().UTC()
	order := &entity.Order{
		ID: uuid.NewString(), UserID: input.UserID, Email: input.Email, Price: input.Price,
		ProductID: input.ProductID, CourierID: input.CourierID, DeliverySlot: input.DeliverySlot,
		Status: entity.StatusPaymentPending, CreatedAt: now, UpdatedAt: now,
	}

	envelope, err := contracts.NewEnvelope(
		contracts.MessageChargePaymentRequested,
		order.ID,
		"",
		contracts.ChargePayment{
			OrderID: order.ID, OperationID: "order:" + order.ID + ":payment",
			UserID: order.UserID, Amount: order.Price,
		},
	)
	if err != nil {
		return nil, err
	}
	payload, err := envelope.Marshal()
	if err != nil {
		return nil, err
	}

	message := &outbox.Message{
		ID: envelope.MessageID, Topic: contracts.TopicBillingCommands,
		MessageKey: order.ID, MessageType: envelope.MessageType, Payload: string(payload),
	}
	requestHash, err := createOrderHash(input)
	if err != nil {
		return nil, err
	}
	return s.repository.CreateIdempotent(ctx, order, message, repositories.IdempotencyRequest{
		UserID: input.UserID, Key: input.IdempotencyKey, RequestHash: requestHash,
	})
}

func createOrderHash(input CreateOrder) (string, error) {
	payload, err := json.Marshal(struct {
		Price        int64  `json:"price"`
		ProductID    string `json:"productId"`
		CourierID    string `json:"courierId"`
		DeliverySlot string `json:"deliverySlot"`
	}{
		Price: input.Price, ProductID: input.ProductID,
		CourierID: input.CourierID, DeliverySlot: input.DeliverySlot,
	})
	if err != nil {
		return "", fmt.Errorf("marshal order payload for idempotency: %w", err)
	}
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:]), nil
}

func (s *OrderService) Get(ctx context.Context, id string, userID int64) (*entity.Order, error) {
	return s.repository.Get(ctx, id, userID)
}

func (s *OrderService) List(ctx context.Context, userID int64) ([]*entity.Order, error) {
	return s.repository.List(ctx, userID)
}
