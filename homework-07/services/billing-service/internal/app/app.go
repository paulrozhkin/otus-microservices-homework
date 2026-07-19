package app

import (
	"context"
	"errors"
	"net/http"

	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/config"
	httpserver "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/repositories"
	"go.uber.org/zap"
)

type App struct {
	cfg    config.Config
	logger *zap.Logger
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	logger, err := newLogger(cfg)
	if err != nil {
		return nil, err
	}
	db, err := platformdb.OpenPostgres(platformdb.PostgresConfig{Host: cfg.DB.Host, Port: cfg.DB.Port, User: cfg.DB.User.Value(), Password: cfg.DB.Password.Value(), DBName: cfg.DB.DBName, SSLMode: cfg.DB.SSLMode, TimeZone: cfg.DB.TimeZone})
	if err != nil {
		return nil, err
	}
	router := httpserver.NewRouter(httpserver.RouterConfig{Config: cfg, Logger: logger, Repository: repositories.NewBillingRepository(db), HealthChecker: platformdb.NewPostgresHealthChecker(db)})
	return &App{cfg: cfg, logger: logger, server: &http.Server{Addr: cfg.HTTP.Addr, Handler: router, ReadTimeout: cfg.HTTP.ReadTimeout, WriteTimeout: cfg.HTTP.WriteTimeout, IdleTimeout: cfg.HTTP.IdleTimeout}}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		err := a.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()
		return a.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func newLogger(cfg config.Config) (*zap.Logger, error) {
	if cfg.IsProduction() {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
