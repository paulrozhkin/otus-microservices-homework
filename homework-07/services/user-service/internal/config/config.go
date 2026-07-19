package config

import (
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/config"
)

type Config struct {
	config.BaseConfig `mapstructure:",squash"`
	Billing           BillingConfig `mapstructure:"billing" validate:"required"`
}

type BillingConfig struct {
	UserServiceBaseURL         string        `mapstructure:"userServiceBaseURL" validate:"required,http_url"`
	UserServiceResponseTimeout time.Duration `mapstructure:"userServiceResponseTimeout" validate:"required"`
}

func Load() (Config, error) {
	configBuilder := config.NewBuilder[Config]().
		WithAppName("user-service").
		WithConfigPath("./services/user-service/config").
		WithEnvPrefix("OTUS_USER_SERVICE").
		WithBindEnv(map[string]string{
			"billing.userServiceBaseURL":         "BILLING_USER_SERVICE_URL",
			"billing.userServiceResponseTimeout": "BILLING_USER_SERVICE_RESPONSE_TIMEOUT",
		}).
		WithDefaultValues(map[string]any{
			"billing.userServiceBaseURL":         "http://billing-service",
			"billing.userServiceResponseTimeout": 5 * time.Second,
		})
	cfg, err := configBuilder.Load()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
