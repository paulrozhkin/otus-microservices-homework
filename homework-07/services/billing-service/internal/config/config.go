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

func (s SecretString) String() string               { return "****" }
func (s SecretString) MarshalJSON() ([]byte, error) { return json.Marshal("****") }
func (s SecretString) Value() string                { return string(s) }

type Config struct {
	App  AppConfig  `mapstructure:"app" validate:"required"`
	HTTP HTTPConfig `mapstructure:"http" validate:"required"`
	DB   DBConfig   `mapstructure:"db" validate:"required"`
}

type AppConfig struct {
	Name string `mapstructure:"name" validate:"required"`
	Env  string `mapstructure:"env" validate:"oneof=production development"`
}

type HTTPConfig struct {
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

func (c Config) IsProduction() bool { return c.App.Env == ProductionEnv }

func Load() (Config, error) {
	v := viper.New()
	setDefaults(v)
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./services/billing-service/config")
	v.SetEnvPrefix("OTUS_BILLING_SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := bindEnvs(v); err != nil {
		return Config{}, err
	}
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := validator.New().Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", DevelopmentEnv)
	v.SetDefault("app.name", "billing-service")
	v.SetDefault("http.addr", ":8001")
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
		"app.env": "OTUS_BILLING_SERVICE_APP_ENV", "app.name": "OTUS_BILLING_SERVICE_APP_NAME",
		"http.addr": "OTUS_BILLING_SERVICE_HTTP_ADDR", "http.readTimeout": "OTUS_BILLING_SERVICE_HTTP_READ_TIMEOUT",
		"http.writeTimeout": "OTUS_BILLING_SERVICE_HTTP_WRITE_TIMEOUT", "http.idleTimeout": "OTUS_BILLING_SERVICE_HTTP_IDLE_TIMEOUT",
		"http.shutdownTimeout": "OTUS_BILLING_SERVICE_HTTP_SHUTDOWN_TIMEOUT", "db.host": "OTUS_BILLING_SERVICE_DB_HOST",
		"db.port": "OTUS_BILLING_SERVICE_DB_PORT", "db.user": "OTUS_BILLING_SERVICE_DB_USER",
		"db.password": "OTUS_BILLING_SERVICE_DB_PASSWORD", "db.dbName": "OTUS_BILLING_SERVICE_DB_DBNAME",
		"db.sslmode": "OTUS_BILLING_SERVICE_DB_SSLMODE", "db.timeZone": "OTUS_BILLING_SERVICE_DB_TIMEZONE",
	}
	for key, env := range envs {
		if err := v.BindEnv(key, env); err != nil {
			return fmt.Errorf("bind %s: %w", env, err)
		}
	}
	return nil
}
