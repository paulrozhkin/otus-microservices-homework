package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/docs"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/repositories"
	"go.uber.org/zap"
)

type RouterConfig struct {
	Config         config.Config
	Logger         *zap.Logger
	UserRepository repositories.UserRepository
	HealthChecker  platformdb.HealthChecker
	BillingClient  handlers.AccountProvisioner
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	r := platformhttp.NewRouter(cfg.Config.IsProduction(), cfg.Logger)
	platformhttp.RegisterOperationalRoutes(r, cfg.HealthChecker)
	r.GET("/health", platformhttp.Liveness)
	r.GET("/swagger.yaml", swaggerYAML)
	r.GET("/swagger", swaggerUI)

	userHandler := handlers.NewUserHandler(cfg.Logger, cfg.UserRepository, cfg.BillingClient)
	authHandler := handlers.NewAuthHandler(cfg.Logger, cfg.UserRepository, cfg.BillingClient)
	profileHandler := handlers.NewProfileHandler(cfg.Logger, cfg.UserRepository)

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/logout", authHandler.Logout)
	r.GET("/auth", authHandler.Auth)

	api := r.Group("/api/v1")
	api.GET("/profile", profileHandler.Get)
	api.PUT("/profile", profileHandler.Update)
	usersGroup := api.Group("/users")
	usersGroup.POST("", userHandler.Create)
	usersGroup.GET("/:id", userHandler.GetByID)
	usersGroup.GET("", userHandler.GetAll)
	usersGroup.PUT("/:id", userHandler.Update)
	usersGroup.DELETE("/:id", userHandler.Delete)
	return r
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
      window.ui = SwaggerUIBundle({url: "/swagger.yaml", dom_id: "#swagger-ui"});
    };
  </script>
</body>
</html>`))
}
