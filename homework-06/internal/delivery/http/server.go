package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/docs"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/delivery/http/middleware"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/repositories"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
)

type RouterConfig struct {
	Config         config.Config
	Logger         *zap.Logger
	UserRepository repositories.UserRepository
	HealthChecker  repositories.HealthChecker
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	if cfg.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.MetricsMiddleware())
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

	r.GET("/health", liveness) // Обратная совместимость с предыдущим ДЗ
	r.GET("/livez", liveness)
	r.GET("/readyz", readiness(cfg.HealthChecker))
	r.GET("/swagger.yaml", swaggerYAML)
	r.GET("/swagger", swaggerUI)
	// Expose metrics endpoint (separate from application routes)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	userHandler := handlers.NewUserHandler(cfg.Logger, cfg.UserRepository)
	authHandler := handlers.NewAuthHandler(cfg.Logger, cfg.UserRepository)
	profileHandler := handlers.NewProfileHandler(cfg.Logger, cfg.UserRepository)

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/logout", authHandler.Logout)
	r.GET("/auth", authHandler.Auth)

	api := r.Group("/api/v1")
	{
		api.GET("/profile", profileHandler.Get)
		api.PUT("/profile", profileHandler.Update)

		usersGroup := api.Group("/users")
		{
			usersGroup.POST("", userHandler.Create)
			usersGroup.GET("/:id", userHandler.GetByID)
			usersGroup.GET("", userHandler.GetAll)
			usersGroup.PUT("/:id", userHandler.Update)
			usersGroup.DELETE("/:id", userHandler.Delete)
		}
	}

	return r
}

func liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func readiness(healthChecker repositories.HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()

		if err := healthChecker.Ping(ctx); err != nil {
			c.Error(entity.ErrServiceUnavailable)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func swaggerYAML(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", docs.SwaggerYAML)
}

func swaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>User Service Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/swagger.yaml",
        dom_id: "#swagger-ui"
      });
    };
  </script>
</body>
</html>`))
}
