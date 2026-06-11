package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	ProductionEnv  = "production"
	DevelopmentEnv = "development"
)

type Config struct {
	App      AppConfig  `mapstructure:"app" validate:"required"`
	Http     HttpConfig `mapstructure:"http" validate:"required"`
	DBConfig DBConfig   `mapstructure:"db" validate:"required"`
}

type AppConfig struct {
	Name string `mapstructure:"name" validate:"required"`
	Env  string `mapstructure:"env" validate:"oneof=production development"`
}

type HttpConfig struct {
	Addr            string        `mapstructure:"addr" validate:"required,hostname_port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout     time.Duration `mapstructure:"idleTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
}

type DBConfig struct {
	Host     string `mapstructure:"host" validate:"required,hostname"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	DBName   string `mapstructure:"dbName" validate:"required"`
	SSLMode  string `mapstructure:"sslmode" validate:"oneof=disable enable"`
	TimeZone string `mapstructure:"timeZone"`
}

func (c Config) IsProduction() bool {
	return c.App.Env == ProductionEnv
}

func Load() (Config, error) {
	v := viper.New()

	setDefaults(v)
	configureReader(v)

	if err := readConfig(v); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func readConfig(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}

		return fmt.Errorf("failed to read config: %w", err)
	}
	return nil
}

func configureReader(v *viper.Viper) {
	v.SetConfigName("config")
	v.SetConfigType("json")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("OTUS_USER_SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "production")
	v.SetDefault("app.name", "user-service")

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.read_timeout", 5*time.Second)
	v.SetDefault("http.write_timeout", 10*time.Second)
	v.SetDefault("http.idle_timeout", 60*time.Second)
	v.SetDefault("http.shutdown_timeout", 15*time.Second)

	v.SetDefault("db.port", 5432)
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.timeZone", "UTC")
}

func validateConfig(cfg Config) error {
	validate := validator.New()
	return validate.Struct(cfg)
}
