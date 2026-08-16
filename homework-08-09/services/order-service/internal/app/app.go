package app

import (
	"context"
	"net/http"

	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/logging"
	platformkafka "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/kafka"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/config"
	httpdelivery "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/repositories"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg          config.Config
	logger       *zap.Logger
	server       *http.Server
	publisher    platformkafka.Publisher
	outboxWorker *outbox.Worker
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
	orderService := service.NewOrderService(repositories.NewOrderRepository(db, outboxRepository))
	router := httpdelivery.NewRouter(httpdelivery.RouterConfig{
		Config: cfg, Logger: logger, OrderService: orderService,
		HealthChecker: platformdb.NewPostgresHealthChecker(db),
	})

	return &App{
		cfg: cfg, logger: logger, server: platformhttp.New(cfg.Http, router), publisher: publisher,
		outboxWorker: outbox.NewWorker(outboxRepository, publisher, logger, cfg.Outbox.PollInterval),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.publisher.Close()
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return platformhttp.Run(groupCtx, a.server, a.cfg.Http.ShutdownTimeout, a.logger)
	})
	group.Go(func() error {
		return a.outboxWorker.Run(groupCtx)
	})
	return group.Wait()
}
