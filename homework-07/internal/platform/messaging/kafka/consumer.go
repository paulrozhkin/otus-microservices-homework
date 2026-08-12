package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"
)

type Handler func(ctx context.Context, key, value []byte) error

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
	Close() error
}

type ReaderConsumer struct {
	reader *segmentkafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *ReaderConsumer {
	return &ReaderConsumer{reader: segmentkafka.NewReader(segmentkafka.ReaderConfig{
		Brokers:               brokers,
		Topic:                 topic,
		GroupID:               groupID,
		StartOffset:           segmentkafka.FirstOffset,
		WatchPartitionChanges: true,
		MinBytes:              1,
		MaxBytes:              10e6,
	})}
}

func (c *ReaderConsumer) Consume(ctx context.Context, handler Handler) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return nil
			}
			if err = waitForRetry(ctx, time.Second); err != nil {
				return nil
			}
			continue
		}
		if err = handler(ctx, message.Key, message.Value); err != nil {
			return fmt.Errorf("handle kafka message: %w", err)
		}
		if err = c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("commit kafka message: %w", err)
		}
	}
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *ReaderConsumer) Close() error {
	return c.reader.Close()
}
