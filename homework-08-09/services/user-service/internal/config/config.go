package config

import (
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
)

type Config struct {
	config.BaseConfig `mapstructure:",squash"`
	Billing           BillingConfig `mapstructure:"billing" validate:"required"`
}

type BillingConfig struct {
	BaseURL         string        `mapstructure:"baseURL" validate:"required,http_url"`
	ResponseTimeout time.Duration `mapstructure:"responseTimeout" validate:"required"`
}

func Load() (Config, error) {
	configBuilder := config.NewBuilder[Config]().
		WithAppName("user-service").
		WithConfigPath("./services/user-service/config").
		WithEnvPrefix("OTUS_USER_SERVICE").
		WithBindEnv(map[string]string{
			"billing.baseURL":         "BILLING_BASE_URL",
			"billing.responseTimeout": "BILLING_RESPONSE_TIMEOUT",
		}).
		WithDefaultValues(map[string]any{
			"billing.baseURL":         "http://billing-service",
			"billing.responseTimeout": 5 * time.Second,
		})
	cfg, err := configBuilder.Load()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
