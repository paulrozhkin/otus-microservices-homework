package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order struct {
	ID            string `gorm:"primaryKey;size:36"`
	UserID        int64  `gorm:"not null;index"`
	Email         string `gorm:"not null"`
	Price         int64  `gorm:"not null;check:order_price_positive,price > 0"`
	ProductID     string `gorm:"not null;size:255"`
	CourierID     string `gorm:"not null;size:255"`
	DeliverySlot  string `gorm:"not null;size:255"`
	Status        string `gorm:"not null;size:32;index"`
	FailureReason string `gorm:"not null;default:''"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type IdempotencyKey struct {
	UserID      int64  `gorm:"primaryKey;autoIncrement:false"`
	Key         string `gorm:"primaryKey;size:255"`
	RequestHash string `gorm:"not null;size:64"`
	OrderID     string `gorm:"not null;size:36;index"`
	CreatedAt   time.Time
}

type IdempotencyRequest struct {
	UserID      int64
	Key         string
	RequestHash string
}

type OrderRepository interface {
	CreateIdempotent(context.Context, *entity.Order, *outbox.Message, IdempotencyRequest) (*entity.Order, error)
	Update(context.Context, *entity.Order) error
	Get(context.Context, string, int64) (*entity.Order, error)
	List(context.Context, int64) ([]*entity.Order, error)
}

type OrderRepositoryImpl struct {
	db     *gorm.DB
	outbox *outbox.Repository
}

func NewOrderRepository(db *gorm.DB, outboxRepository *outbox.Repository) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{db: db, outbox: outboxRepository}
}

func (r *OrderRepositoryImpl) ApplyPaymentSucceeded(ctx context.Context, orderID, causationID string) error {
	clearReason := ""
	return r.applyTransition(ctx, orderID, entity.StatusPaymentPending, entity.StatusInventoryPending, &clearReason,
		func(order *Order) (*outbox.Message, error) { return newReserveInventoryMessage(order, causationID) })
}

func newReserveInventoryMessage(order *Order, causationID string) (*outbox.Message, error) {
	return newCommandMessage(contracts.TopicWarehouseCommands, contracts.MessageReserveInventoryRequested, order, causationID,
		contracts.ReserveInventory{OrderID: order.ID, OperationID: "order:" + order.ID + ":inventory:reserve", ProductID: order.ProductID})
}

func (r *OrderRepositoryImpl) ApplyPaymentFailed(ctx context.Context, orderID, reason string) error {
	if reason == "" {
		reason = "payment failed"
	}
	return r.applyTransition(ctx, orderID, entity.StatusPaymentPending, entity.StatusFailed, &reason, nil)
}

func (r *OrderRepositoryImpl) ApplyInventoryReserved(ctx context.Context, orderID, causationID string) error {
	return r.applyTransition(ctx, orderID, entity.StatusInventoryPending, entity.StatusDeliveryPending, nil,
		func(order *Order) (*outbox.Message, error) { return newReserveDeliveryMessage(order, causationID) })
}

func (r *OrderRepositoryImpl) ApplyInventoryReservationFailed(ctx context.Context, orderID, reason, causationID string) error {
	return r.applyTransition(ctx, orderID, entity.StatusInventoryPending, entity.StatusPaymentRefunding, &reason,
		func(order *Order) (*outbox.Message, error) { return newRefundPaymentMessage(order, causationID) })
}

func (r *OrderRepositoryImpl) ApplyDeliveryReserved(ctx context.Context, orderID, causationID string) error {
	return r.applyTransition(ctx, orderID, entity.StatusDeliveryPending, entity.StatusCompleted, nil,
		func(order *Order) (*outbox.Message, error) { return newNotificationMessage(order, causationID) })
}

func (r *OrderRepositoryImpl) ApplyDeliveryReservationFailed(ctx context.Context, orderID, reason, causationID string) error {
	return r.applyTransition(ctx, orderID, entity.StatusDeliveryPending, entity.StatusInventoryReleasing, &reason,
		func(order *Order) (*outbox.Message, error) { return newReleaseInventoryMessage(order, causationID) })
}

func (r *OrderRepositoryImpl) ApplyInventoryReleased(ctx context.Context, orderID, causationID string) error {
	return r.applyTransition(ctx, orderID, entity.StatusInventoryReleasing, entity.StatusPaymentRefunding, nil,
		func(order *Order) (*outbox.Message, error) { return newRefundPaymentMessage(order, causationID) })
}

func (r *OrderRepositoryImpl) ApplyPaymentRefunded(ctx context.Context, orderID string) error {
	return r.applyTransition(ctx, orderID, entity.StatusPaymentRefunding, entity.StatusFailed, nil, nil)
}

type transitionMessageBuilder func(*Order) (*outbox.Message, error)

func (r *OrderRepositoryImpl) applyTransition(
	ctx context.Context,
	orderID string,
	expectedStatus, targetStatus entity.Status,
	failureReason *string,
	buildMessage transitionMessageBuilder,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order := &Order{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(order, "id = ?", orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: order %s", apperror.ErrNotFound, orderID)
			}
			return err
		}
		if entity.Status(order.Status) != expectedStatus {
			return nil
		}

		var message *outbox.Message
		var err error
		if buildMessage != nil {
			message, err = buildMessage(order)
			if err != nil {
				return err
			}
		}

		updates := map[string]any{"status": string(targetStatus)}
		if failureReason != nil {
			updates["failure_reason"] = *failureReason
		}
		if err = tx.Model(&Order{}).Where("id = ?", order.ID).Updates(updates).Error; err != nil {
			return err
		}
		if message != nil {
			return r.outbox.Enqueue(ctx, tx, message)
		}
		return nil
	})
}

func newReserveDeliveryMessage(order *Order, causationID string) (*outbox.Message, error) {
	return newCommandMessage(contracts.TopicDeliveryCommands, contracts.MessageReserveDeliveryRequested, order, causationID,
		contracts.ReserveDelivery{OrderID: order.ID, OperationID: "order:" + order.ID + ":delivery:reserve", CourierID: order.CourierID, DeliverySlot: order.DeliverySlot})
}

func newReleaseInventoryMessage(order *Order, causationID string) (*outbox.Message, error) {
	return newCommandMessage(contracts.TopicWarehouseCommands, contracts.MessageReleaseInventoryRequested, order, causationID,
		contracts.ReleaseInventory{OrderID: order.ID, OperationID: "order:" + order.ID + ":inventory:release", ProductID: order.ProductID})
}

func newRefundPaymentMessage(order *Order, causationID string) (*outbox.Message, error) {
	return newCommandMessage(contracts.TopicBillingCommands, contracts.MessageRefundPaymentRequested, order, causationID,
		contracts.RefundPayment{
			OrderID: order.ID, OperationID: "order:" + order.ID + ":payment:refund",
			OriginalOperationID: "order:" + order.ID + ":payment",
		})
}

func newNotificationMessage(order *Order, causationID string) (*outbox.Message, error) {
	return newCommandMessage(contracts.TopicNotificationCommands, contracts.MessageNotificationRequested, order, causationID,
		contracts.SendNotification{
			OrderID: order.ID, UserID: order.UserID, Email: order.Email,
			OrderStatus: string(entity.StatusCompleted), Subject: "Order completed",
			Body: fmt.Sprintf("Order %s has been completed", order.ID),
		})
}

func newCommandMessage(topic, messageType string, order *Order, causationID string, command any) (*outbox.Message, error) {
	envelope, err := contracts.NewEnvelope(messageType, order.ID, causationID, command)
	if err != nil {
		return nil, err
	}
	payload, err := envelope.Marshal()
	if err != nil {
		return nil, err
	}
	return &outbox.Message{
		ID: envelope.MessageID, Topic: topic, MessageKey: order.ID,
		MessageType: envelope.MessageType, Payload: string(payload),
	}, nil
}

func (r *OrderRepositoryImpl) CreateIdempotent(
	ctx context.Context,
	order *entity.Order,
	message *outbox.Message,
	request IdempotencyRequest,
) (*entity.Order, error) {
	var result *entity.Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		idempotencyKey := &IdempotencyKey{
			UserID: request.UserID, Key: request.Key, RequestHash: request.RequestHash, OrderID: order.ID,
		}
		insert := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(idempotencyKey)
		if insert.Error != nil {
			return insert.Error
		}
		if insert.RowsAffected == 0 {
			existingKey := &IdempotencyKey{}
			if err := tx.First(existingKey, "user_id = ? AND key = ?", request.UserID, request.Key).Error; err != nil {
				return err
			}
			if existingKey.RequestHash != request.RequestHash {
				return fmt.Errorf("%w: Idempotency-Key was already used with another payload", apperror.ErrAlreadyExists)
			}

			existingOrder := &Order{}
			if err := tx.First(existingOrder, "id = ? AND user_id = ?", existingKey.OrderID, request.UserID).Error; err != nil {
				return err
			}
			result = orderFromDAO(existingOrder)
			return nil
		}

		dao := orderToDAO(order)
		if err := tx.Create(dao).Error; err != nil {
			return err
		}
		if err := r.outbox.Enqueue(ctx, tx, message); err != nil {
			return err
		}
		copyOrder(order, dao)
		result = order
		return nil
	})
	return result, err
}

func (r *OrderRepositoryImpl) Update(ctx context.Context, order *entity.Order) error {
	dao := orderToDAO(order)
	result := r.db.WithContext(ctx).Model(&Order{}).Where("id = ? AND user_id = ?", order.ID, order.UserID).Updates(map[string]any{
		"status": dao.Status, "failure_reason": dao.FailureReason,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: order %s", apperror.ErrNotFound, order.ID)
	}
	return nil
}

func (r *OrderRepositoryImpl) Get(ctx context.Context, id string, userID int64) (*entity.Order, error) {
	dao := &Order{}
	if err := r.db.WithContext(ctx).First(dao, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: order %s", apperror.ErrNotFound, id)
		}
		return nil, err
	}
	return orderFromDAO(dao), nil
}

func (r *OrderRepositoryImpl) List(ctx context.Context, userID int64) ([]*entity.Order, error) {
	var daos []*Order
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&daos).Error; err != nil {
		return nil, err
	}
	orders := make([]*entity.Order, 0, len(daos))
	for _, dao := range daos {
		orders = append(orders, orderFromDAO(dao))
	}
	return orders, nil
}

func orderToDAO(order *entity.Order) *Order {
	return &Order{
		ID: order.ID, UserID: order.UserID, Email: order.Email, Price: order.Price,
		ProductID: order.ProductID, CourierID: order.CourierID, DeliverySlot: order.DeliverySlot,
		Status: string(order.Status), FailureReason: order.FailureReason,
		CreatedAt: order.CreatedAt, UpdatedAt: order.UpdatedAt,
	}
}

func orderFromDAO(dao *Order) *entity.Order {
	return &entity.Order{
		ID: dao.ID, UserID: dao.UserID, Email: dao.Email, Price: dao.Price,
		ProductID: dao.ProductID, CourierID: dao.CourierID, DeliverySlot: dao.DeliverySlot,
		Status: entity.Status(dao.Status), FailureReason: dao.FailureReason,
		CreatedAt: dao.CreatedAt, UpdatedAt: dao.UpdatedAt,
	}
}

func copyOrder(order *entity.Order, dao *Order) {
	order.CreatedAt, order.UpdatedAt = dao.CreatedAt, dao.UpdatedAt
}
