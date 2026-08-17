package app

import (
	"context"
	"net/http"

	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/logging"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/clients"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/config"
	httpserver "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/repositories"
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

	logger.Info("Configuration initialized", zap.Any("config", cfg))

	db, err := platformdb.OpenPostgres(cfg.DBConfig)
	if err != nil {
		return nil, err
	}

	router := httpserver.NewRouter(httpserver.RouterConfig{
		Config:         cfg,
		Logger:         logger,
		UserRepository: repositories.NewUserRepository(db),
		HealthChecker:  platformdb.NewPostgresHealthChecker(db),
		BillingClient:  clients.NewBillingClient(cfg.Billing.BaseURL, cfg.Billing.ResponseTimeout),
	})
	server := platformhttp.New(cfg.Http, router)

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting HTTP server", zap.String("addr", a.cfg.Http.Addr))
	return platformhttp.Run(ctx, a.server, a.cfg.Http.ShutdownTimeout, a.logger)
}
