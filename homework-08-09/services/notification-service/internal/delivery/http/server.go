package http

import (
	"github.com/gin-gonic/gin"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/docs"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/repositories"
	"go.uber.org/zap"
)

type RouterConfig struct {
	Config        config.Config
	Logger        *zap.Logger
	Repository    repositories.NotificationRepository
	HealthChecker platformdb.HealthChecker
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	r := platformhttp.NewRouter(cfg.Config.IsProduction(), cfg.Logger)
	platformhttp.RegisterOperationalRoutes(r, cfg.HealthChecker)
	platformhttp.RegisterSwaggerRoutes(r, "Notification Service", docs.SwaggerYAML)
	h := handlers.NewNotificationHandler(cfg.Repository)
	r.GET("/api/v1/notifications", h.List)
	return r
}
