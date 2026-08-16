package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ReservationReserved = "reserved"
	OperationSucceeded  = "succeeded"
	OperationFailed     = "failed"
)

type Reservation struct {
	OrderID      string `gorm:"primaryKey;size:36"`
	CourierID    string `gorm:"not null;size:255;uniqueIndex:uidx_active_delivery,where:status = 'reserved'"`
	DeliverySlot string `gorm:"not null;size:255;uniqueIndex:uidx_active_delivery,where:status = 'reserved'"`
	Status       string `gorm:"not null;size:32;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type Operation struct {
	ID           string `gorm:"primaryKey;size:128"`
	OrderID      string `gorm:"not null;size:36;index"`
	CourierID    string `gorm:"not null;size:255"`
	DeliverySlot string `gorm:"not null;size:255"`
	Status       string `gorm:"not null;size:32"`
	Reason       string `gorm:"not null;default:''"`
	CreatedAt    time.Time
}
type DeliveryRepository struct {
	db     *gorm.DB
	outbox *outbox.Repository
}

func NewDeliveryRepository(db *gorm.DB, outboxRepository *outbox.Repository) *DeliveryRepository {
	return &DeliveryRepository{db: db, outbox: outboxRepository}
}
func (r *DeliveryRepository) Reserve(ctx context.Context, command contracts.ReserveDelivery, succeeded, failed *outbox.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing := &Operation{}
		if err := tx.First(existing, "id = ?", command.OperationID).Error; err == nil {
			if existing.OrderID != command.OrderID || existing.CourierID != command.CourierID || existing.DeliverySlot != command.DeliverySlot {
				return fmt.Errorf("operation %s has different parameters", command.OperationID)
			}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		reservation := &Reservation{OrderID: command.OrderID, CourierID: command.CourierID, DeliverySlot: command.DeliverySlot, Status: ReservationReserved}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(reservation)
		if result.Error != nil {
			return result.Error
		}
		status, reason, message := OperationSucceeded, "", succeeded
		if result.RowsAffected == 0 {
			status, reason, message = OperationFailed, "courier slot is already reserved", failed
		}
		if err := tx.Create(&Operation{ID: command.OperationID, OrderID: command.OrderID, CourierID: command.CourierID, DeliverySlot: command.DeliverySlot, Status: status, Reason: reason}).Error; err != nil {
			return err
		}
		return r.outbox.Enqueue(ctx, tx, message)
	})
}
