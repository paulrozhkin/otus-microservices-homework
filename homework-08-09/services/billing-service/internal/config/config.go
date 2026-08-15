package config

import (
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
)

type Config struct {
	config.BaseConfig `mapstructure:",squash"`
}

func Load() (Config, error) {
	configBuilder := config.NewBuilder[Config]().
		WithAppName("billing-service").
		WithConfigPath("./services/billing-service/config").
		WithEnvPrefix("OTUS_BILLING_SERVICE")
	cfg, err := configBuilder.Load()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}
