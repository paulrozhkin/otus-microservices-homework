package config

import (
	"encoding/json"
	"time"
)

const (
	ProductionEnv  = "production"
	DevelopmentEnv = "development"
)

type BaseConfig struct {
	App      AppConfig  `mapstructure:"app" validate:"required"`
	Http     HttpConfig `mapstructure:"http" validate:"required"`
	DBConfig DBConfig   `mapstructure:"db" validate:"required"`
}

func (c BaseConfig) IsProduction() bool {
	return c.App.Env == ProductionEnv
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

// DBConfig contains connection settings shared by all services.
type DBConfig struct {
	Host     string       `mapstructure:"host" validate:"required,hostname"`
	Port     int          `mapstructure:"port"`
	User     SecretString `mapstructure:"user" validate:"required"`
	Password SecretString `mapstructure:"password" validate:"required"`
	DBName   string       `mapstructure:"dbName" validate:"required"`
	SSLMode  string       `mapstructure:"sslmode" validate:"oneof=disable enable"`
	TimeZone string       `mapstructure:"timeZone"`
}

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
