package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/app"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/repositories"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	mode := "serve"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	isMigration := false
	switch mode {
	case "serve":
	case "migrate":
		isMigration = true
	default:
		log.Fatalf("unknown mode: %s", mode)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if isMigration {
		err = runMigrations(cfg)
		if err != nil {
			log.Fatalf("failed to migrate db due to: %v", err)
		}
		log.Println("migrations completed")
		return
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}

func runMigrations(cfg config.Config) error {
	db, err := repositories.NewDbConnection(cfg.DBConfig)
	if err != nil {
		return err
	}
	if err = db.AutoMigrate(&repositories.User{}, &repositories.Session{}); err != nil {
		return err
	}
	return nil
}
