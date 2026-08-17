package main

import (
	"context"
	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/database"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/messaging/outbox"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/delivery-service/internal/app"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/delivery-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/delivery-service/internal/repositories"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		db, e := platformdb.OpenPostgres(cfg.DBConfig)
		if e == nil {
			e = db.AutoMigrate(&repositories.Reservation{}, &repositories.Operation{}, &outbox.Message{})
		}
		if e != nil {
			log.Fatalf("migrate: %v", e)
		}
		return
	}
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("initialize app: %v", err)
	}
	if err = application.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
