package app

import (
	"context"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/logging"
	platformkafka "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/kafka"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/delivery-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/delivery-service/internal/consumer"
	httpdelivery "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/delivery-service/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/delivery-service/internal/repositories"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type App struct {
	cfg       config.Config
	logger    *zap.Logger
	server    *http.Server
	consumer  platformkafka.Consumer
	handler   platformkafka.Handler
	publisher platformkafka.Publisher
	worker    *outbox.Worker
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
	publisher := platformkafka.NewPublisher(cfg.Kafka.BrokerList())
	outboxRepository := outbox.NewRepository(db)
	repository := repositories.NewDeliveryRepository(db, outboxRepository)
	handler := consumer.NewHandler(repository)
	reader := platformkafka.NewConsumer(cfg.Kafka.BrokerList(), cfg.Kafka.Topic, cfg.Kafka.GroupID)
	router := httpdelivery.NewRouter(cfg, logger, platformdb.NewPostgresHealthChecker(db))
	return &App{cfg: cfg, logger: logger, server: platformhttp.New(cfg.Http, router), consumer: reader, handler: handler.Handle, publisher: publisher, worker: outbox.NewWorker(outboxRepository, publisher, logger, cfg.Outbox.PollInterval)}, nil
}
func (a *App) Run(ctx context.Context) error {
	defer a.consumer.Close()
	defer a.publisher.Close()
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return platformhttp.Run(groupCtx, a.server, a.cfg.Http.ShutdownTimeout, a.logger) })
	group.Go(func() error { return a.consumer.Consume(groupCtx, a.handler) })
	group.Go(func() error { return a.worker.Run(groupCtx) })
	return group.Wait()
}
