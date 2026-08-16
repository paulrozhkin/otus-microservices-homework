package config

import (
	"strings"
	"time"

	platformconfig "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
)

type Config struct {
	platformconfig.BaseConfig `mapstructure:",squash"`
	Kafka                     KafkaConfig  `mapstructure:"kafka" validate:"required"`
	Outbox                    OutboxConfig `mapstructure:"outbox" validate:"required"`
}

type KafkaConfig struct {
	Brokers string `mapstructure:"brokers" validate:"required"`
}

type OutboxConfig struct {
	PollInterval time.Duration `mapstructure:"pollInterval" validate:"required"`
}

func (c KafkaConfig) BrokerList() []string {
	parts := strings.Split(c.Brokers, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		if broker := strings.TrimSpace(part); broker != "" {
			brokers = append(brokers, broker)
		}
	}
	return brokers
}

func Load() (Config, error) {
	cfg, err := platformconfig.NewBuilder[Config]().
		WithAppName("order-service").
		WithConfigPath("./services/order-service/config").
		WithEnvPrefix("OTUS_ORDER_SERVICE").
		WithBindEnv(map[string]string{
			"kafka.brokers":       "KAFKA_BROKERS",
			"outbox.pollInterval": "OUTBOX_POLL_INTERVAL",
		}).
		WithDefaultValues(map[string]any{
			"http.addr":           ":8002",
			"kafka.brokers":       "kafka:29092",
			"outbox.pollInterval": 500 * time.Millisecond,
		}).Load()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
