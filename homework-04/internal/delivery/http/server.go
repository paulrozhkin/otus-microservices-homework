package http

import (
	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/delivery/http/middleware"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/repositories"
	"go.uber.org/zap"
)

type RouterConfig struct {
	Config         config.Config
	Logger         *zap.Logger
	UserRepository repositories.UserRepository
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	if cfg.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestID())

	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		cfg.Logger.Info("http request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("request_id", param.Keys["request_id"].(string)),
		)
		return ""
	}))
	r.Use(middleware.ErrorHandler(cfg.Logger))

	userHandler := handlers.NewUserHandler(cfg.Logger, cfg.UserRepository)

	api := r.Group("/api/v1")
	{
		usersGroup := api.Group("/users")
		{
			usersGroup.POST("", userHandler.Create)
			usersGroup.GET("/:id", userHandler.GetByID)
			usersGroup.GET("", userHandler.GetAll)
			usersGroup.PATCH("/:id", userHandler.Update)
			usersGroup.DELETE("/:id", userHandler.Delete)
		}
	}

	return r
}
