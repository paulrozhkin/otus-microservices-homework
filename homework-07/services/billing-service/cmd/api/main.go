package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	platformdb "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/database"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/app"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/repositories"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	mode := "serve"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if mode == "migrate" {
		if err = migrate(cfg); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		return
	}
	if mode != "serve" {
		log.Fatalf("unknown mode: %s", mode)
	}
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("initialize app: %v", err)
	}
	if err = application.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}

func migrate(cfg config.Config) error {
	db, err := platformdb.OpenPostgres(platformdb.PostgresConfig{Host: cfg.DB.Host, Port: cfg.DB.Port, User: cfg.DB.User.Value(), Password: cfg.DB.Password.Value(), DBName: cfg.DB.DBName, SSLMode: cfg.DB.SSLMode, TimeZone: cfg.DB.TimeZone})
	if err != nil {
		return err
	}
	return db.AutoMigrate(&repositories.Account{}, &repositories.Operation{})
}
