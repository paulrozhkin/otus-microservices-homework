package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/events"
	businessmetrics "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/metrics"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/repositories"
)

type Billing interface {
	Pay(context.Context, string, int64, int64) error
}
type Publisher interface {
	Publish(context.Context, []byte, []byte) error
}

type OrderService struct {
	repository repositories.OrderRepository
	billing    Billing
	publisher  Publisher
}

func NewOrderService(repository repositories.OrderRepository, billing Billing, publisher Publisher) *OrderService {
	return &OrderService{repository: repository, billing: billing, publisher: publisher}
}

func (s *OrderService) Create(ctx context.Context, userID int64, email string, price int64) (*entity.Order, error) {
	if price <= 0 {
		return nil, fmt.Errorf("%w: price must be positive", apperror.ErrInvalidOperation)
	}
	now := time.Now().UTC()
	order := &entity.Order{ID: uuid.NewString(), UserID: userID, Email: email, Price: price, Status: entity.StatusPending, CreatedAt: now, UpdatedAt: now}
	if err := s.repository.Create(ctx, order); err != nil {
		return nil, err
	}

	err := s.billing.Pay(ctx, order.ID, order.UserID, order.Price)
	switch {
	case err == nil:
		order.Status = entity.StatusPaid
	case errors.Is(err, apperror.ErrInsufficientFunds):
		order.Status = entity.StatusRejected
		order.FailureReason = apperror.ErrInsufficientFunds.Error()
	default:
		order.Status = entity.StatusFailed
		order.FailureReason = "billing unavailable"
		_ = s.repository.Update(ctx, order)
		return nil, err
	}
	order.UpdatedAt = time.Now().UTC()
	if err = s.repository.Update(ctx, order); err != nil {
		return nil, err
	}
	businessmetrics.OrderFinalized(string(order.Status), order.Price)
	if err = s.publishNotification(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) Get(ctx context.Context, id string, userID int64) (*entity.Order, error) {
	return s.repository.Get(ctx, id, userID)
}

func (s *OrderService) List(ctx context.Context, userID int64) ([]*entity.Order, error) {
	return s.repository.List(ctx, userID)
}

func (s *OrderService) publishNotification(ctx context.Context, order *entity.Order) error {
	event := events.NotificationRequested{
		EventID: uuid.NewString(), EventType: events.NotificationRequestedV1,
		OrderID: order.ID, UserID: order.UserID, Email: order.Email,
		OrderStatus: string(order.Status), OccurredAt: time.Now().UTC(),
	}
	if order.Status == entity.StatusPaid {
		event.Subject, event.Body = "Order paid", "Your order has been successfully paid"
	} else {
		event.Subject, event.Body = "Order rejected", "Your order was rejected: "+order.FailureReason
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal notification event: %w", err)
	}
	if err = s.publisher.Publish(ctx, []byte(order.ID), payload); err != nil {
		return fmt.Errorf("publish notification event: %w", err)
	}
	return nil
}
