package database

import (
	"context"
	"fmt"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenPostgres(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User.Value(), cfg.Password.Value(), cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone,
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
