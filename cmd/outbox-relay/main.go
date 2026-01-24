package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Niiaks/Aegis/internal/config"
	"github.com/Niiaks/Aegis/internal/database"
	"github.com/Niiaks/Aegis/internal/kafka"
	"github.com/Niiaks/Aegis/internal/logger"
	"github.com/Niiaks/Aegis/internal/outbox"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	loggerService := logger.New(cfg.Observability)
	defer loggerService.Shutdown()
	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	log.Info().Msg("Starting Outbox Relay Service...")

	db, err := database.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer db.Close()

	kProducer, err := kafka.NewProducer(kafka.DefaultConfig(cfg.Kafka.Brokers), &log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed produce")
	}
	defer kProducer.Close()

	relay := outbox.NewRelay(db.Pool, kProducer, &log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := relay.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Relay service stopped with error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down Outbox Relay...")
	cancel()

	log.Info().Msg("Outbox Relay shutdown complete")
}
