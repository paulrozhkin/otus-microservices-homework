package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Builder[T any] struct {
	configPath    string
	envPrefix     string
	appName       string
	defaultValues map[string]any
	bindEnv       map[string]string
}

func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{}
}

func (b *Builder[T]) WithConfigPath(s string) *Builder[T] {
	b.configPath = s
	return b
}

func (b *Builder[T]) WithEnvPrefix(s string) *Builder[T] {
	b.envPrefix = s
	return b
}

func (b *Builder[T]) WithDefaultValues(defaultValues map[string]any) *Builder[T] {
	b.defaultValues = defaultValues
	return b
}

func (b *Builder[T]) WithBindEnv(bindEnv map[string]string) *Builder[T] {
	b.bindEnv = bindEnv
	return b
}

func (b *Builder[T]) WithAppName(appName string) *Builder[T] {
	b.appName = appName
	return b
}

func (b *Builder[T]) Load() (*T, error) {
	v := viper.New()

	b.setDefaults(v)

	if err := b.configureReader(v); err != nil {
		return nil, err
	}

	if err := readConfig(v); err != nil {
		return nil, err
	}

	cfg := new(T)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func (b *Builder[T]) configureReader(v *viper.Viper) error {
	v.SetConfigName("config")
	v.SetConfigType("json")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if b.configPath != "" {
		v.AddConfigPath(b.configPath)
	}

	if b.envPrefix != "" {
		v.SetEnvPrefix(b.envPrefix)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return b.bindEnvs(v)
}

func (b *Builder[T]) setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "production")
	if b.appName != "" {
		v.SetDefault("app.name", b.appName)
	} else {
		v.SetDefault("app.name", "default")
	}

	v.SetDefault("http.addr", ":8000")
	v.SetDefault("http.readTimeout", 5*time.Second)
	v.SetDefault("http.writeTimeout", 10*time.Second)
	v.SetDefault("http.idleTimeout", 60*time.Second)
	v.SetDefault("http.shutdownTimeout", 15*time.Second)

	v.SetDefault("db.port", 5432)
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.timeZone", "UTC")
	v.SetDefault("db.startupTimeout", 30*time.Second)
	v.SetDefault("db.retryInterval", time.Second)

	if b.defaultValues != nil {
		for key, value := range b.defaultValues {
			v.SetDefault(key, value)
		}
	}
}

func (b *Builder[T]) addPrefixToEnv(env string) string {
	if b.envPrefix != "" {
		return fmt.Sprintf("%s_%s", b.envPrefix, env)
	}
	return env
}

func (b *Builder[T]) bindEnvs(v *viper.Viper) error {
	envs := map[string]string{
		"app.env":              b.addPrefixToEnv("APP_ENV"),
		"app.name":             b.addPrefixToEnv("APP_NAME"),
		"http.addr":            b.addPrefixToEnv("HTTP_ADDR"),
		"http.readTimeout":     b.addPrefixToEnv("HTTP_READ_TIMEOUT"),
		"http.writeTimeout":    b.addPrefixToEnv("HTTP_WRITE_TIMEOUT"),
		"http.idleTimeout":     b.addPrefixToEnv("HTTP_IDLE_TIMEOUT"),
		"http.shutdownTimeout": b.addPrefixToEnv("HTTP_SHUTDOWN_TIMEOUT"),
		"db.host":              b.addPrefixToEnv("DB_HOST"),
		"db.port":              b.addPrefixToEnv("DB_PORT"),
		"db.user":              b.addPrefixToEnv("DB_USER"),
		"db.password":          b.addPrefixToEnv("DB_PASSWORD"),
		"db.dbName":            b.addPrefixToEnv("DB_DBNAME"),
		"db.sslmode":           b.addPrefixToEnv("DB_SSLMODE"),
		"db.timeZone":          b.addPrefixToEnv("DB_TIMEZONE"),
		"db.startupTimeout":    b.addPrefixToEnv("DB_STARTUP_TIMEOUT"),
		"db.retryInterval":     b.addPrefixToEnv("DB_RETRY_INTERVAL"),
	}

	if b.bindEnv != nil {
		for key, value := range b.bindEnv {
			envs[key] = b.addPrefixToEnv(value)
		}
	}

	for key, env := range envs {
		if err := v.BindEnv(key, env); err != nil {
			return fmt.Errorf("failed to bind env %s: %w", env, err)
		}
	}

	return nil
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

func validateConfig(cfg any) error {
	validate := validator.New()
	return validate.Struct(cfg)
}
