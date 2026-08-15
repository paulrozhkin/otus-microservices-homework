package repositories

import (
	"context"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Notification struct {
	EventID     string `gorm:"primaryKey;size:36"`
	OrderID     string `gorm:"not null;size:36;index"`
	UserID      int64  `gorm:"not null;index"`
	Email       string `gorm:"not null"`
	OrderStatus string `gorm:"not null;size:32"`
	Subject     string `gorm:"not null"`
	Body        string `gorm:"not null"`
	CreatedAt   time.Time
}

type NotificationRepository interface {
	Create(context.Context, *entity.Notification) error
	List(context.Context, int64) ([]*entity.Notification, error)
}

type NotificationRepositoryImpl struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &NotificationRepositoryImpl{db: db}
}

func (r *NotificationRepositoryImpl) Create(ctx context.Context, notification *entity.Notification) error {
	dao := notificationToDAO(notification)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(dao).Error
}

func (r *NotificationRepositoryImpl) List(ctx context.Context, userID int64) ([]*entity.Notification, error) {
	var daos []*Notification
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&daos).Error; err != nil {
		return nil, err
	}
	notifications := make([]*entity.Notification, 0, len(daos))
	for _, dao := range daos {
		notifications = append(notifications, notificationFromDAO(dao))
	}
	return notifications, nil
}

func notificationToDAO(notification *entity.Notification) *Notification {
	return &Notification{
		EventID: notification.EventID, OrderID: notification.OrderID, UserID: notification.UserID,
		Email: notification.Email, OrderStatus: notification.OrderStatus, Subject: notification.Subject,
		Body: notification.Body, CreatedAt: notification.CreatedAt,
	}
}

func notificationFromDAO(dao *Notification) *entity.Notification {
	return &entity.Notification{
		EventID: dao.EventID, OrderID: dao.OrderID, UserID: dao.UserID,
		Email: dao.Email, OrderStatus: dao.OrderStatus, Subject: dao.Subject,
		Body: dao.Body, CreatedAt: dao.CreatedAt,
	}
}
