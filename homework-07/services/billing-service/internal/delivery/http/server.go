package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	platformmiddleware "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/httpmiddleware"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/delivery/http/middleware"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/repositories"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type RouterConfig struct {
	Config        config.Config
	Logger        *zap.Logger
	Repository    repositories.BillingRepository
	HealthChecker platformdb.HealthChecker
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	if cfg.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(platformmiddleware.Metrics(), platformmiddleware.RequestID(), gin.Recovery(), middleware.ErrorHandler(cfg.Logger))
	h := handlers.NewBillingHandler(cfg.Repository)
	r.GET("/livez", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", readiness(cfg.HealthChecker))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.PUT("/internal/v1/accounts/:userId", h.CreateAccount)
	r.POST("/internal/v1/payments", h.Pay)
	r.POST("/internal/v1/payments/:operationId/refund", h.Refund)
	api := r.Group("/api/v1/billing")
	api.GET("/account", h.GetAccount)
	api.POST("/deposits", h.Deposit)
	api.POST("/withdrawals", h.Withdraw)
	return r
}

func readiness(checker platformdb.HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if checker == nil || checker.Ping(ctx) != nil {
			c.Error(entity.ErrServiceUnavailable)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
