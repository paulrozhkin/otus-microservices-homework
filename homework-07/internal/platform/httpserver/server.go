package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/httpmiddleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func NewRouter(production bool, logger *zap.Logger, mappings ...httpmiddleware.ErrorMapping) *gin.Engine {
	if production {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(
		httpmiddleware.Metrics(),
		httpmiddleware.RequestID(),
		gin.Recovery(),
		httpmiddleware.RequestLogger(logger),
		httpmiddleware.ErrorHandler(logger, mappings...),
	)
	return r
}

func RegisterOperationalRoutes(r *gin.Engine, checker database.HealthChecker) {
	r.GET("/livez", Liveness)
	r.GET("/readyz", Readiness(checker))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func Readiness(checker database.HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if checker == nil || checker.Ping(ctx) != nil {
			c.Error(apperror.ErrServiceUnavailable)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func New(cfg config.HttpConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr: cfg.Addr, Handler: handler,
		ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout,
	}
}

func Run(ctx context.Context, server *http.Server, shutdownTimeout time.Duration, logger *zap.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown HTTP server", zap.Error(err))
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
