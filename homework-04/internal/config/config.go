package config

import (
	"encoding/json"
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

type SecretString string

func (s SecretString) String() string {
	return "****"
}

func (s SecretString) MarshalJSON() ([]byte, error) {
	return json.Marshal("****")
}

func (s SecretString) Value() string {
	return string(s)
}

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
	Host     string       `mapstructure:"host" validate:"required,hostname"`
	Port     int          `mapstructure:"port"`
	User     SecretString `mapstructure:"user" validate:"required"`
	Password SecretString `mapstructure:"password" validate:"required"`
	DBName   string       `mapstructure:"dbName" validate:"required"`
	SSLMode  string       `mapstructure:"sslmode" validate:"oneof=disable enable"`
	TimeZone string       `mapstructure:"timeZone"`
}

func (c Config) IsProduction() bool {
	return c.App.Env == ProductionEnv
}

func Load() (Config, error) {
	v := viper.New()

	setDefaults(v)
	if err := configureReader(v); err != nil {
		return Config{}, err
	}

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

func configureReader(v *viper.Viper) error {
	v.SetConfigName("config")
	v.SetConfigType("json")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("OTUS_USER_SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return bindEnvs(v)
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "production")
	v.SetDefault("app.name", "users-service")

	v.SetDefault("http.addr", ":8000")
	v.SetDefault("http.readTimeout", 5*time.Second)
	v.SetDefault("http.writeTimeout", 10*time.Second)
	v.SetDefault("http.idleTimeout", 60*time.Second)
	v.SetDefault("http.shutdownTimeout", 15*time.Second)

	v.SetDefault("db.port", 5432)
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.timeZone", "UTC")
}

func bindEnvs(v *viper.Viper) error {
	envs := map[string]string{
		"app.env":              "OTUS_USER_SERVICE_APP_ENV",
		"app.name":             "OTUS_USER_SERVICE_APP_NAME",
		"http.addr":            "OTUS_USER_SERVICE_HTTP_ADDR",
		"http.readTimeout":     "OTUS_USER_SERVICE_HTTP_READ_TIMEOUT",
		"http.writeTimeout":    "OTUS_USER_SERVICE_HTTP_WRITE_TIMEOUT",
		"http.idleTimeout":     "OTUS_USER_SERVICE_HTTP_IDLE_TIMEOUT",
		"http.shutdownTimeout": "OTUS_USER_SERVICE_HTTP_SHUTDOWN_TIMEOUT",
		"db.host":              "OTUS_USER_SERVICE_DB_HOST",
		"db.port":              "OTUS_USER_SERVICE_DB_PORT",
		"db.user":              "OTUS_USER_SERVICE_DB_USER",
		"db.password":          "OTUS_USER_SERVICE_DB_PASSWORD",
		"db.dbName":            "OTUS_USER_SERVICE_DB_DBNAME",
		"db.sslmode":           "OTUS_USER_SERVICE_DB_SSLMODE",
		"db.timeZone":          "OTUS_USER_SERVICE_DB_TIMEZONE",
	}

	for key, env := range envs {
		if err := v.BindEnv(key, env); err != nil {
			return fmt.Errorf("failed to bind env %s: %w", env, err)
		}
	}

	return nil
}

func validateConfig(cfg Config) error {
	validate := validator.New()
	return validate.Struct(cfg)
}
