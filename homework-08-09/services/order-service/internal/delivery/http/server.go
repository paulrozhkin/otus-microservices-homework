package http

import (
	"github.com/gin-gonic/gin"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/docs"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/service"
	"go.uber.org/zap"
)

type RouterConfig struct {
	Config        config.Config
	Logger        *zap.Logger
	OrderService  *service.OrderService
	HealthChecker platformdb.HealthChecker
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	r := platformhttp.NewRouter(cfg.Config.IsProduction(), cfg.Logger)
	platformhttp.RegisterOperationalRoutes(r, cfg.HealthChecker)
	platformhttp.RegisterSwaggerRoutes(r, "Order Service", docs.SwaggerYAML)
	h := handlers.NewOrderHandler(cfg.OrderService)
	api := r.Group("/api/v1/orders")
	api.POST("", h.Create)
	api.GET("", h.List)
	api.GET("/:id", h.Get)
	return r
}
