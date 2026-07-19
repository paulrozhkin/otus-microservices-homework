package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/entity"
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
	ID        string `gorm:"primaryKey;size:128"`
	UserID    int64  `gorm:"not null;index"`
	Type      string `gorm:"not null;size:32"`
	Amount    int64  `gorm:"not null;check:amount_positive,amount > 0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BillingRepository interface {
	CreateAccount(context.Context, int64) (*entity.Account, bool, error)
	GetAccount(context.Context, int64) (*entity.Account, error)
	Credit(context.Context, string, int64, int64, string) (*entity.Account, error)
	Debit(context.Context, string, int64, int64, string) (*entity.Account, error)
	Refund(context.Context, string) (*entity.Account, error)
}

type BillingRepositoryImpl struct{ db *gorm.DB }

func NewBillingRepository(db *gorm.DB) BillingRepository { return &BillingRepositoryImpl{db: db} }

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
		if err := tx.Create(&Operation{ID: operationID, UserID: userID, Type: operationType, Amount: amount}).Error; err != nil {
			return err
		}
		result = accountFromDAO(account)
		return nil
	})
	return result, err
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
	return r.Credit(ctx, "refund:"+originalOperationID, original.UserID, original.Amount, entity.OperationRefund)
}

func accountFromDAO(account *Account) *entity.Account {
	return &entity.Account{UserID: account.UserID, Balance: account.Balance, CreatedAt: account.CreatedAt, UpdatedAt: account.UpdatedAt}
}
