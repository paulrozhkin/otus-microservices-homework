package app

import (
	"context"
	"net/http"

	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/logging"
	platformkafka "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/kafka"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/config"
	eventsconsumer "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/consumer"
	httpdelivery "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/repositories"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg      config.Config
	logger   *zap.Logger
	server   *http.Server
	consumer platformkafka.Consumer
	handler  platformkafka.Handler
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
	repository := repositories.NewNotificationRepository(db)
	handler := eventsconsumer.NewNotificationHandler(repository)
	kafkaConsumer := platformkafka.NewConsumer(cfg.Kafka.BrokerList(), cfg.Kafka.Topic, cfg.Kafka.GroupID, logger)
	router := httpdelivery.NewRouter(httpdelivery.RouterConfig{
		Config: cfg, Logger: logger, Repository: repository,
		HealthChecker: platformdb.NewPostgresHealthChecker(db),
	})
	return &App{
		cfg: cfg, logger: logger, server: platformhttp.New(cfg.Http, router),
		consumer: kafkaConsumer, handler: handler.Handle,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return platformhttp.Run(groupCtx, a.server, a.cfg.Http.ShutdownTimeout, a.logger)
	})
	group.Go(func() error {
		return a.consumer.Consume(groupCtx, a.handler)
	})
	err := group.Wait()
	if closeErr := a.consumer.Close(); err == nil {
		err = closeErr
	}
	return err
}
