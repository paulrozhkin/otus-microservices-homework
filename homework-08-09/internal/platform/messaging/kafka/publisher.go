package kafka

import (
	"context"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"
)

type Publisher interface {
	Publish(ctx context.Context, key, value []byte) error
	Close() error
}

type WriterPublisher struct {
	writer *segmentkafka.Writer
}

func NewPublisher(brokers []string, topic string) *WriterPublisher {
	return &WriterPublisher{writer: &segmentkafka.Writer{
		Addr:                   segmentkafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &segmentkafka.Hash{},
		RequiredAcks:           segmentkafka.RequireAll,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}}
}

func (p *WriterPublisher) Publish(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, segmentkafka.Message{
		Key: key, Value: value, Time: time.Now().UTC(),
	})
}

func (p *WriterPublisher) Close() error {
	return p.writer.Close()
}
