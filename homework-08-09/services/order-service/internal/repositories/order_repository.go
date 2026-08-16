package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/entity"
	"gorm.io/gorm"
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

type OrderRepository interface {
	Create(context.Context, *entity.Order, *outbox.Message) error
	Update(context.Context, *entity.Order) error
	Get(context.Context, string, int64) (*entity.Order, error)
	List(context.Context, int64) ([]*entity.Order, error)
}

type OrderRepositoryImpl struct {
	db     *gorm.DB
	outbox *outbox.Repository
}

func NewOrderRepository(db *gorm.DB, outboxRepository *outbox.Repository) OrderRepository {
	return &OrderRepositoryImpl{db: db, outbox: outboxRepository}
}

func (r *OrderRepositoryImpl) Create(ctx context.Context, order *entity.Order, message *outbox.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dao := orderToDAO(order)
		if err := tx.Create(dao).Error; err != nil {
			return err
		}
		if err := r.outbox.Enqueue(ctx, tx, message); err != nil {
			return err
		}
		copyOrder(order, dao)
		return nil
	})
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
