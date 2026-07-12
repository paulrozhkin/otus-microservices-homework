package repositories

import (
	"context"
	"fmt"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type DBHealthChecker struct {
	db *gorm.DB
}

func NewDbConnection(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User.Value(), cfg.Password.Value(),
		cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func NewDBHealthChecker(db *gorm.DB) *DBHealthChecker {
	return &DBHealthChecker{db: db}
}

func (h *DBHealthChecker) Ping(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}
