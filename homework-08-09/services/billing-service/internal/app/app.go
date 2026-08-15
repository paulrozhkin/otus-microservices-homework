package app

import (
	"context"
	"net/http"

	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/logging"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/billing-service/internal/config"
	httpserver "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/billing-service/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/billing-service/internal/repositories"
	"go.uber.org/zap"
)

type App struct {
	cfg    config.Config
	logger *zap.Logger
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	logger, err := logging.New(cfg.IsProduction())
	if err != nil {
		return nil, err
	}
	db, err := platformdb.OpenPostgres(cfg.DBConfig)
	if err != nil {
		return nil, err
	}
	router := httpserver.NewRouter(httpserver.RouterConfig{Config: cfg, Logger: logger, Repository: repositories.NewBillingRepository(db), HealthChecker: platformdb.NewPostgresHealthChecker(db)})
	return &App{cfg: cfg, logger: logger, server: platformhttp.New(cfg.Http, router)}, nil
}

func (a *App) Run(ctx context.Context) error {
	return platformhttp.Run(ctx, a.server, a.cfg.Http.ShutdownTimeout, a.logger)
}
