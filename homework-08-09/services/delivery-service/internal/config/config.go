package config

import (
	"strings"
	"time"

	platformconfig "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/contracts"
)

type Config struct {
	platformconfig.BaseConfig `mapstructure:",squash"`
	Kafka                     KafkaConfig  `mapstructure:"kafka" validate:"required"`
	Outbox                    OutboxConfig `mapstructure:"outbox" validate:"required"`
}
type KafkaConfig struct {
	Brokers string `mapstructure:"brokers" validate:"required"`
	Topic   string `mapstructure:"topic" validate:"required"`
	GroupID string `mapstructure:"groupId" validate:"required"`
}
type OutboxConfig struct {
	PollInterval time.Duration `mapstructure:"pollInterval" validate:"required"`
}

func (c KafkaConfig) BrokerList() []string {
	var result []string
	for _, value := range strings.Split(c.Brokers, ",") {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
func Load() (Config, error) {
	cfg, err := platformconfig.NewBuilder[Config]().WithAppName("delivery-service").WithConfigPath("./services/delivery-service/config").WithEnvPrefix("OTUS_DELIVERY_SERVICE").
		WithBindEnv(map[string]string{"kafka.brokers": "KAFKA_BROKERS", "kafka.topic": "KAFKA_TOPIC", "kafka.groupId": "KAFKA_GROUP_ID", "outbox.pollInterval": "OUTBOX_POLL_INTERVAL"}).
		WithDefaultValues(map[string]any{"http.addr": ":8005", "kafka.brokers": "kafka:29092", "kafka.topic": contracts.TopicDeliveryCommands, "kafka.groupId": "delivery-service-v1", "outbox.pollInterval": 500 * time.Millisecond}).Load()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
