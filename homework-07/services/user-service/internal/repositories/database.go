package repositories

import (
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/config"
	"gorm.io/gorm"
)

type HealthChecker = platformdb.HealthChecker
type DBHealthChecker = platformdb.PostgresHealthChecker

func NewDbConnection(cfg config.DBConfig) (*gorm.DB, error) {
	return platformdb.OpenPostgres(platformdb.PostgresConfig{
		Host: cfg.Host, Port: cfg.Port, User: cfg.User.Value(), Password: cfg.Password.Value(),
		DBName: cfg.DBName, SSLMode: cfg.SSLMode, TimeZone: cfg.TimeZone,
	})
}

func NewDBHealthChecker(db *gorm.DB) *DBHealthChecker {
	return platformdb.NewPostgresHealthChecker(db)
}
