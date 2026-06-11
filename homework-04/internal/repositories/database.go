package repositories

import (
	"fmt"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDbConnection(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password,
		cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
