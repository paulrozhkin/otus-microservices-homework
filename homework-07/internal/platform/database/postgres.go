package database

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfig contains connection settings shared by all services.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

func OpenPostgres(cfg PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone,
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

type HealthChecker interface {
	Ping(context.Context) error
}

type PostgresHealthChecker struct {
	db *gorm.DB
}

func NewPostgresHealthChecker(db *gorm.DB) *PostgresHealthChecker {
	return &PostgresHealthChecker{db: db}
}

func (h *PostgresHealthChecker) Ping(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}
