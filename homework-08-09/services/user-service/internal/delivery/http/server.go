package http

import (
	"github.com/gin-gonic/gin"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/user-service/docs"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/user-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/user-service/internal/delivery/http/handlers"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/user-service/internal/repositories"
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
	platformhttp.RegisterSwaggerRoutes(r, "User Service", docs.SwaggerYAML)

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
