package http

import (
	"github.com/gin-gonic/gin"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/warehouse-service/internal/config"
	"go.uber.org/zap"
)

func NewRouter(cfg config.Config, logger *zap.Logger, healthChecker platformdb.HealthChecker) *gin.Engine {
	router := platformhttp.NewRouter(cfg.IsProduction(), logger)
	platformhttp.RegisterOperationalRoutes(router, healthChecker)
	return router
}
