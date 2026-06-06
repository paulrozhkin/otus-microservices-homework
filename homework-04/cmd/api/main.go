package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/app"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}
