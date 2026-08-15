package config

import (
	"strings"

	platformconfig "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
)

type Config struct {
	platformconfig.BaseConfig `mapstructure:",squash"`
	Kafka                     KafkaConfig `mapstructure:"kafka" validate:"required"`
}

type KafkaConfig struct {
	Brokers string `mapstructure:"brokers" validate:"required"`
	Topic   string `mapstructure:"topic" validate:"required"`
	GroupID string `mapstructure:"groupId" validate:"required"`
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
		WithAppName("notification-service").
		WithConfigPath("./services/notification-service/config").
		WithEnvPrefix("OTUS_NOTIFICATION_SERVICE").
		WithBindEnv(map[string]string{
			"kafka.brokers": "KAFKA_BROKERS",
			"kafka.topic":   "KAFKA_TOPIC",
			"kafka.groupId": "KAFKA_GROUP_ID",
		}).
		WithDefaultValues(map[string]any{
			"http.addr":     ":8003",
			"kafka.brokers": "kafka:29092",
			"kafka.topic":   "notification.requested.v1",
			"kafka.groupId": "notification-service-v1",
		}).Load()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
