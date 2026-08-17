package config

import (
	"strings"
	"time"

	platformconfig "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/config"
)

type Config struct {
	platformconfig.BaseConfig `mapstructure:",squash"`
	Billing                   BillingConfig `mapstructure:"billing" validate:"required"`
	Kafka                     KafkaConfig   `mapstructure:"kafka" validate:"required"`
}

type BillingConfig struct {
	BaseURL         string        `mapstructure:"baseURL" validate:"required,http_url"`
	ResponseTimeout time.Duration `mapstructure:"responseTimeout" validate:"required"`
}

type KafkaConfig struct {
	Brokers string `mapstructure:"brokers" validate:"required"`
	Topic   string `mapstructure:"topic" validate:"required"`
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
			"billing.baseURL":         "BILLING_BASE_URL",
			"billing.responseTimeout": "BILLING_RESPONSE_TIMEOUT",
			"kafka.brokers":           "KAFKA_BROKERS",
			"kafka.topic":             "KAFKA_TOPIC",
		}).
		WithDefaultValues(map[string]any{
			"http.addr":               ":8002",
			"billing.baseURL":         "http://billing-service:8001",
			"billing.responseTimeout": 5 * time.Second,
			"kafka.brokers":           "kafka:29092",
			"kafka.topic":             "notification.requested.v1",
		}).Load()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
