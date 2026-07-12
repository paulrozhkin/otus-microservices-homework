package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/config"
	httpserver "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/repositories"
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

	logger.Info("Configuration initialized", zap.Any("config", cfg))

	db, err := repositories.NewDbConnection(cfg.DBConfig)
	if err != nil {
		return nil, err
	}

	router := httpserver.NewRouter(httpserver.RouterConfig{
		Config:         cfg,
		Logger:         logger,
		UserRepository: repositories.NewUserRepository(db),
		HealthChecker:  repositories.NewDBHealthChecker(db),
	})

	server := &http.Server{
		Addr:         cfg.Http.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Http.ReadTimeout,
		WriteTimeout: cfg.Http.WriteTimeout,
		IdleTimeout:  cfg.Http.IdleTimeout,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting HTTP server", zap.String("addr", a.cfg.Http.Addr))

	errCh := make(chan error, 1)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.Http.ShutdownTimeout)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("failed to shutdown HTTP server", zap.Error(err))
			return err
		}

		a.logger.Info("HTTP server stopped gracefully")
		return nil

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
