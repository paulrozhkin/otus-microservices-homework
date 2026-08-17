package kafka

import (
	"context"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"
)

type Publisher interface {
	Publish(ctx context.Context, message Message) error
	Close() error
}

type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

type WriterPublisher struct {
	writer *segmentkafka.Writer
}

func NewPublisher(brokers []string) *WriterPublisher {
	return &WriterPublisher{writer: &segmentkafka.Writer{
		Addr:                   segmentkafka.TCP(brokers...),
		Balancer:               &segmentkafka.Hash{},
		RequiredAcks:           segmentkafka.RequireAll,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}}
}

func (p *WriterPublisher) Publish(ctx context.Context, message Message) error {
	return p.writer.WriteMessages(ctx, segmentkafka.Message{
		Topic: message.Topic,
		Key:   message.Key,
		Value: message.Value,
		Time:  time.Now().UTC(),
	})
}

func (p *WriterPublisher) Close() error {
	return p.writer.Close()
}
