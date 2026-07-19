package http

import (
	"github.com/gin-gonic/gin"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/repositories"
	"go.uber.org/zap"
)

type RouterConfig struct {
	Config        config.Config
	Logger        *zap.Logger
	Repository    repositories.BillingRepository
	HealthChecker platformdb.HealthChecker
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	r := platformhttp.NewRouter(cfg.Config.IsProduction(), cfg.Logger)
	platformhttp.RegisterOperationalRoutes(r, cfg.HealthChecker)
	h := handlers.NewBillingHandler(cfg.Repository)
	r.PUT("/internal/v1/accounts/:userId", h.CreateAccount)
	r.POST("/internal/v1/payments", h.Pay)
	r.POST("/internal/v1/payments/:operationId/refund", h.Refund)
	api := r.Group("/api/v1/billing")
	api.GET("/account", h.GetAccount)
	api.POST("/deposits", h.Deposit)
	api.POST("/withdrawals", h.Withdraw)
	return r
}
