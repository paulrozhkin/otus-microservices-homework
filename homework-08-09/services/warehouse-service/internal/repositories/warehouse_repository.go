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
	ReservationReleased = "released"
	OperationReserve    = "reserve"
	OperationRelease    = "release"
	OperationSucceeded  = "succeeded"
	OperationFailed     = "failed"
)

type Reservation struct {
	OrderID   string `gorm:"primaryKey;size:36"`
	ProductID string `gorm:"not null;size:255;uniqueIndex:uidx_active_product,where:status = 'reserved'"`
	Status    string `gorm:"not null;size:32;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
type Operation struct {
	ID        string `gorm:"primaryKey;size:128"`
	OrderID   string `gorm:"not null;size:36;index"`
	ProductID string `gorm:"not null;size:255"`
	Type      string `gorm:"not null;size:32"`
	Status    string `gorm:"not null;size:32"`
	Reason    string `gorm:"not null;default:''"`
	CreatedAt time.Time
}
type WarehouseRepository struct {
	db     *gorm.DB
	outbox *outbox.Repository
}

func NewWarehouseRepository(db *gorm.DB, outboxRepository *outbox.Repository) *WarehouseRepository {
	return &WarehouseRepository{db: db, outbox: outboxRepository}
}

func (r *WarehouseRepository) Reserve(ctx context.Context, command contracts.ReserveInventory, succeeded, failed *outbox.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing := &Operation{}
		if err := tx.First(existing, "id = ?", command.OperationID).Error; err == nil {
			if existing.OrderID != command.OrderID || existing.ProductID != command.ProductID || existing.Type != OperationReserve {
				return fmt.Errorf("operation %s has different parameters", command.OperationID)
			}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		reservation := &Reservation{OrderID: command.OrderID, ProductID: command.ProductID, Status: ReservationReserved}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(reservation)
		if result.Error != nil {
			return result.Error
		}
		status, reason, message := OperationSucceeded, "", succeeded
		if result.RowsAffected == 0 {
			status, reason, message = OperationFailed, "product is already reserved", failed
		}
		if err := tx.Create(&Operation{ID: command.OperationID, OrderID: command.OrderID, ProductID: command.ProductID, Type: OperationReserve, Status: status, Reason: reason}).Error; err != nil {
			return err
		}
		return r.outbox.Enqueue(ctx, tx, message)
	})
}

func (r *WarehouseRepository) Release(ctx context.Context, command contracts.ReleaseInventory, released *outbox.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing := &Operation{}
		if err := tx.First(existing, "id = ?", command.OperationID).Error; err == nil {
			if existing.OrderID != command.OrderID || existing.ProductID != command.ProductID || existing.Type != OperationRelease {
				return fmt.Errorf("operation %s has different parameters", command.OperationID)
			}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		reservation := &Reservation{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(reservation, "order_id = ?", command.OrderID).Error; err != nil {
			return err
		}
		if reservation.ProductID != command.ProductID {
			return fmt.Errorf("order %s reserved another product", command.OrderID)
		}
		if reservation.Status == ReservationReserved {
			if err := tx.Model(reservation).Update("status", ReservationReleased).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&Operation{ID: command.OperationID, OrderID: command.OrderID, ProductID: command.ProductID, Type: OperationRelease, Status: OperationSucceeded}).Error; err != nil {
			return err
		}
		return r.outbox.Enqueue(ctx, tx, released)
	})
}
