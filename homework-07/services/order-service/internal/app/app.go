package app

import (
	"context"
	"net/http"

	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	platformhttp "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/httpserver"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/logging"
	platformkafka "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/messaging/kafka"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/order-service/internal/clients"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/order-service/internal/config"
	httpdelivery "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/order-service/internal/delivery/http"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/order-service/internal/repositories"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/order-service/internal/service"
	"go.uber.org/zap"
)

type App struct {
	cfg       config.Config
	logger    *zap.Logger
	server    *http.Server
	publisher platformkafka.Publisher
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
	publisher := platformkafka.NewPublisher(cfg.Kafka.BrokerList(), cfg.Kafka.Topic)
	orderService := service.NewOrderService(
		repositories.NewOrderRepository(db),
		clients.NewBillingClient(cfg.Billing.BaseURL, cfg.Billing.ResponseTimeout),
		publisher,
	)
	router := httpdelivery.NewRouter(httpdelivery.RouterConfig{Config: cfg, Logger: logger, OrderService: orderService, HealthChecker: platformdb.NewPostgresHealthChecker(db)})
	return &App{cfg: cfg, logger: logger, server: platformhttp.New(cfg.Http, router), publisher: publisher}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.publisher.Close()
	return platformhttp.Run(ctx, a.server, a.cfg.Http.ShutdownTimeout, a.logger)
}
