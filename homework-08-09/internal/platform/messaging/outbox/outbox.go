package outbox

import (
	"context"
	"fmt"
	"time"

	platformkafka "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Message struct {
	ID          string     `gorm:"primaryKey;size:36"`
	Topic       string     `gorm:"not null;size:255;index"`
	MessageKey  string     `gorm:"not null;size:255"`
	MessageType string     `gorm:"not null;size:255;index"`
	Payload     string     `gorm:"not null;type:jsonb"`
	Attempts    int        `gorm:"not null;default:0"`
	LastError   string     `gorm:"not null;default:''"`
	PublishedAt *time.Time `gorm:"index"`
	CreatedAt   time.Time  `gorm:"not null;index"`
}

func (Message) TableName() string { return "outbox_messages" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Enqueue(ctx context.Context, tx *gorm.DB, message *Message) error {
	if tx == nil {
		tx = r.db
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}
	if err := tx.WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("enqueue outbox message: %w", err)
	}
	return nil
}

func (r *Repository) PublishNext(ctx context.Context, publisher platformkafka.Publisher) (bool, error) {
	var found bool
	var publishErr error

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		message := &Message{}
		result := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("published_at IS NULL").
			Order("created_at ASC").
			Limit(1).
			Find(message)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		found = true
		publishErr = publisher.Publish(ctx, platformkafka.Message{
			Topic: message.Topic,
			Key:   []byte(message.MessageKey),
			Value: []byte(message.Payload),
		})
		if publishErr != nil {
			return tx.Model(&Message{}).Where("id = ?", message.ID).Updates(map[string]any{
				"attempts":   gorm.Expr("attempts + 1"),
				"last_error": publishErr.Error(),
			}).Error
		}

		now := time.Now().UTC()
		return tx.Model(&Message{}).Where("id = ?", message.ID).Updates(map[string]any{
			"published_at": now,
			"last_error":   "",
		}).Error
	})
	if err != nil {
		return found, fmt.Errorf("publish next outbox message: %w", err)
	}
	if publishErr != nil {
		return found, fmt.Errorf("publish outbox message: %w", publishErr)
	}
	return found, nil
}

type Worker struct {
	repository   *Repository
	publisher    platformkafka.Publisher
	logger       *zap.Logger
	pollInterval time.Duration
}

func NewWorker(repository *Repository, publisher platformkafka.Publisher, logger *zap.Logger, pollInterval time.Duration) *Worker {
	return &Worker{repository: repository, publisher: publisher, logger: logger, pollInterval: pollInterval}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		published, err := w.repository.PublishNext(ctx, w.publisher)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.logger.Warn("outbox publish failed", zap.Error(err))
			if err = wait(ctx, w.pollInterval); err != nil {
				return nil
			}
			continue
		}
		if published {
			continue
		}
		if err = wait(ctx, w.pollInterval); err != nil {
			return nil
		}
	}
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
