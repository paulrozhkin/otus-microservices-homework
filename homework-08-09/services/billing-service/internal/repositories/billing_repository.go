package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/billing-service/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Account struct {
	UserID    int64 `gorm:"primaryKey;autoIncrement:false"`
	Balance   int64 `gorm:"not null;default:0;check:balance_non_negative,balance >= 0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Operation struct {
	ID          string `gorm:"primaryKey;size:128"`
	UserID      int64  `gorm:"not null;index"`
	Type        string `gorm:"not null;size:32"`
	Amount      int64  `gorm:"not null;check:amount_positive,amount > 0"`
	Status      string `gorm:"not null;size:32;default:succeeded"`
	ReferenceID string `gorm:"not null;size:128;default:''"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BillingRepository interface {
	CreateAccount(context.Context, int64) (*entity.Account, bool, error)
	GetAccount(context.Context, int64) (*entity.Account, error)
	Credit(context.Context, string, int64, int64, string) (*entity.Account, error)
	Debit(context.Context, string, int64, int64, string) (*entity.Account, error)
	Refund(context.Context, string) (*entity.Account, error)
}

const (
	OperationStatusSucceeded = "succeeded"
	OperationStatusFailed    = "failed"
)

type BillingRepositoryImpl struct {
	db     *gorm.DB
	outbox *outbox.Repository
}

func NewBillingRepository(db *gorm.DB, outboxRepository ...*outbox.Repository) *BillingRepositoryImpl {
	repository := &BillingRepositoryImpl{db: db}
	if len(outboxRepository) > 0 {
		repository.outbox = outboxRepository[0]
	}
	return repository
}

func (r *BillingRepositoryImpl) CreateAccount(ctx context.Context, userID int64) (*entity.Account, bool, error) {
	account := &Account{UserID: userID}
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).FirstOrCreate(account)
	if result.Error != nil {
		return nil, false, result.Error
	}
	return accountFromDAO(account), result.RowsAffected == 1, nil
}

func (r *BillingRepositoryImpl) GetAccount(ctx context.Context, userID int64) (*entity.Account, error) {
	account := &Account{}
	if err := r.db.WithContext(ctx).First(account, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: account for user %d", apperror.ErrNotFound, userID)
		}
		return nil, err
	}
	return accountFromDAO(account), nil
}

func (r *BillingRepositoryImpl) Credit(ctx context.Context, operationID string, userID, amount int64, operationType string) (*entity.Account, error) {
	return r.apply(ctx, operationID, userID, amount, operationType, false)
}

func (r *BillingRepositoryImpl) Debit(ctx context.Context, operationID string, userID, amount int64, operationType string) (*entity.Account, error) {
	return r.apply(ctx, operationID, userID, amount, operationType, true)
}

func (r *BillingRepositoryImpl) apply(ctx context.Context, operationID string, userID, amount int64, operationType string, debit bool) (*entity.Account, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", apperror.ErrInvalidOperation)
	}
	var result *entity.Account
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing := &Operation{}
		if err := tx.First(existing, "id = ?", operationID).Error; err == nil {
			if existing.UserID != userID || existing.Amount != amount || existing.Type != operationType {
				return fmt.Errorf("%w: operation id %s was already used with different parameters", apperror.ErrInvalidOperation, operationID)
			}
			if existing.Status == OperationStatusFailed {
				return apperror.ErrInsufficientFunds
			}
			account := &Account{}
			if err = tx.First(account, "user_id = ?", userID).Error; err != nil {
				return err
			}
			result = accountFromDAO(account)
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		account := &Account{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(account, "user_id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: account for user %d", apperror.ErrNotFound, userID)
			}
			return err
		}
		if debit {
			if account.Balance < amount {
				return apperror.ErrInsufficientFunds
			}
			account.Balance -= amount
		} else {
			account.Balance += amount
		}
		if err := tx.Save(account).Error; err != nil {
			return err
		}
		if err := tx.Create(&Operation{ID: operationID, UserID: userID, Type: operationType, Amount: amount, Status: OperationStatusSucceeded}).Error; err != nil {
			return err
		}
		result = accountFromDAO(account)
		return nil
	})
	return result, err
}

// ProcessPayment atomically records the business result and its Saga event.
// A failed operation is persisted as well, so replaying the same command can
// never turn a previously rejected payment into a successful one.
func (r *BillingRepositoryImpl) ProcessPayment(ctx context.Context, operationID string, userID, amount int64, succeeded, failed *outbox.Message) error {
	if r.outbox == nil {
		return errors.New("outbox repository is required")
	}
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", apperror.ErrInvalidOperation)
	}
	if succeeded == nil || failed == nil {
		return errors.New("payment result outbox messages are required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing := &Operation{}
		if err := tx.First(existing, "id = ?", operationID).Error; err == nil {
			if existing.UserID != userID || existing.Amount != amount || existing.Type != entity.OperationPayment {
				return fmt.Errorf("%w: operation id %s was already used with different parameters", apperror.ErrInvalidOperation, operationID)
			}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		account := &Account{}
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(account, "user_id = ?", userID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		status := OperationStatusSucceeded
		message := succeeded
		if errors.Is(err, gorm.ErrRecordNotFound) || account.Balance < amount {
			status = OperationStatusFailed
			message = failed
		} else {
			account.Balance -= amount
			if err = tx.Save(account).Error; err != nil {
				return err
			}
		}

		if err = tx.Create(&Operation{ID: operationID, UserID: userID, Type: entity.OperationPayment, Amount: amount, Status: status}).Error; err != nil {
			return err
		}
		return r.outbox.Enqueue(ctx, tx, message)
	})
}

// ProcessRefund atomically credits a previously successful payment and emits
// the compensation result. Replays are no-ops after the first commit.
func (r *BillingRepositoryImpl) ProcessRefund(ctx context.Context, operationID, originalOperationID string, refunded *outbox.Message) error {
	if r.outbox == nil {
		return errors.New("outbox repository is required")
	}
	if refunded == nil {
		return errors.New("refund result outbox message is required")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing := &Operation{}
		if err := tx.First(existing, "id = ?", operationID).Error; err == nil {
			if existing.Type != entity.OperationRefund || existing.ReferenceID != originalOperationID {
				return fmt.Errorf("%w: operation id %s was already used", apperror.ErrInvalidOperation, operationID)
			}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		original := &Operation{}
		if err := tx.First(original, "id = ?", originalOperationID).Error; err != nil {
			return err
		}
		if original.Type != entity.OperationPayment || original.Status != OperationStatusSucceeded {
			return fmt.Errorf("%w: operation %s cannot be refunded", apperror.ErrInvalidOperation, originalOperationID)
		}

		account := &Account{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(account, "user_id = ?", original.UserID).Error; err != nil {
			return err
		}
		anotherRefund := &Operation{}
		if err := tx.First(anotherRefund, "type = ? AND reference_id = ?", entity.OperationRefund, originalOperationID).Error; err == nil {
			return fmt.Errorf("%w: operation %s was already refunded", apperror.ErrInvalidOperation, originalOperationID)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		account.Balance += original.Amount
		if err := tx.Save(account).Error; err != nil {
			return err
		}
		if err := tx.Create(&Operation{ID: operationID, UserID: original.UserID, Type: entity.OperationRefund, Amount: original.Amount, Status: OperationStatusSucceeded, ReferenceID: originalOperationID}).Error; err != nil {
			return err
		}
		return r.outbox.Enqueue(ctx, tx, refunded)
	})
}

func (r *BillingRepositoryImpl) Refund(ctx context.Context, originalOperationID string) (*entity.Account, error) {
	original := &Operation{}
	if err := r.db.WithContext(ctx).First(original, "id = ?", originalOperationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: operation %s", apperror.ErrNotFound, originalOperationID)
		}
		return nil, err
	}
	if original.Type != entity.OperationPayment && original.Type != entity.OperationWithdrawal {
		return nil, fmt.Errorf("%w: operation %s cannot be refunded", apperror.ErrInvalidOperation, originalOperationID)
	}
	if original.Status == OperationStatusFailed {
		return nil, fmt.Errorf("%w: operation %s cannot be refunded", apperror.ErrInvalidOperation, originalOperationID)
	}
	return r.Credit(ctx, "refund:"+originalOperationID, original.UserID, original.Amount, entity.OperationRefund)
}

func accountFromDAO(account *Account) *entity.Account {
	return &entity.Account{UserID: account.UserID, Balance: account.Balance, CreatedAt: account.CreatedAt, UpdatedAt: account.UpdatedAt}
}
