package kafka

import (
	"context"
	"errors"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Handler func(ctx context.Context, key, value []byte) error

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
	Close() error
}

type ReaderConsumer struct {
	reader  *segmentkafka.Reader
	logger  *zap.Logger
	topic   string
	groupID string
}

func NewConsumer(brokers []string, topic, groupID string, loggers ...*zap.Logger) *ReaderConsumer {
	logger := zap.NewNop()
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	}
	return &ReaderConsumer{reader: segmentkafka.NewReader(segmentkafka.ReaderConfig{
		Brokers:               brokers,
		Topic:                 topic,
		GroupID:               groupID,
		StartOffset:           segmentkafka.FirstOffset,
		WatchPartitionChanges: true,
		MinBytes:              1,
		MaxBytes:              10e6,
	}), logger: logger, topic: topic, groupID: groupID}
}

func (c *ReaderConsumer) Consume(ctx context.Context, handler Handler) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return nil
			}
			c.logger.Warn("kafka fetch failed",
				zap.String("topic", c.topic), zap.String("consumer_group", c.groupID), zap.Error(err))
			if err = waitForRetry(ctx, time.Second); err != nil {
				return nil
			}
			continue
		}
		for {
			if err = handler(ctx, message.Key, message.Value); err == nil {
				break
			}
			c.logger.Error("kafka message handling failed",
				zap.String("topic", message.Topic), zap.String("consumer_group", c.groupID),
				zap.String("message_key", string(message.Key)), zap.Int("partition", message.Partition),
				zap.Int64("offset", message.Offset), zap.Error(err))
			if err = waitForRetry(ctx, time.Second); err != nil {
				return nil
			}
		}
		for {
			if err = c.reader.CommitMessages(ctx, message); err == nil {
				break
			}
			c.logger.Warn("kafka commit failed",
				zap.String("topic", message.Topic), zap.String("consumer_group", c.groupID),
				zap.String("message_key", string(message.Key)), zap.Int("partition", message.Partition),
				zap.Int64("offset", message.Offset), zap.Error(err))
			if err = waitForRetry(ctx, time.Second); err != nil {
				return nil
			}
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
