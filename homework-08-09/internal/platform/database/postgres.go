package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenPostgres(cfg config.DBConfig) (*gorm.DB, error) {
	startupTimeout := cfg.StartupTimeout
	if startupTimeout <= 0 {
		startupTimeout = 30 * time.Second
	}
	retryInterval := cfg.RetryInterval
	if retryInterval <= 0 {
		retryInterval = time.Second
	}
	attemptTimeout := min(startupTimeout, 5*time.Second)
	connectTimeoutSeconds := max(1, int((attemptTimeout+time.Second-1)/time.Second))
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s connect_timeout=%d",
		cfg.Host, cfg.User.Value(), cfg.Password.Value(), cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone, connectTimeoutSeconds,
	)

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return openPostgresWithRetry(dsn, address, startupTimeout, retryInterval, func(dsn string) (*gorm.DB, error) {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	})
}

type postgresOpener func(string) (*gorm.DB, error)

func openPostgresWithRetry(dsn, address string, startupTimeout, retryInterval time.Duration, open postgresOpener) (*gorm.DB, error) {
	deadline := time.Now().Add(startupTimeout)
	var lastErr error
	attempts := 0

	for {
		attempts++
		db, err := open(dsn)
		if err == nil {
			return db, nil
		}
		lastErr = err
		if db != nil {
			if sqlDB, dbErr := db.DB(); dbErr == nil {
				_ = sqlDB.Close()
			}
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("connect to postgres at %s after %d attempts within %s: %w", address, attempts, startupTimeout, lastErr)
		}
		wait := min(retryInterval, remaining)
		log.Printf("postgres at %s is not ready (attempt %d); retrying in %s", address, attempts, wait.Round(time.Millisecond))
		timer := time.NewTimer(wait)
		<-timer.C
	}
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
