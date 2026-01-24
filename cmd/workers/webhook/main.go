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
	"github.com/Niiaks/Aegis/internal/redis"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	loggerService := logger.New(cfg.Observability)
	defer loggerService.Shutdown()
	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	log.Info().Msg("Starting Webhook Worker...")

	db, err := database.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer db.Close()

	redis, err := redis.New(&log, &cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize redis")
	}
	defer redis.Close()

	// Initialize Webhook Worker
	consumer, err := kafka.NewConsumer(kafka.DefaultConfig(cfg.Kafka.Brokers), kafka.GroupWebhookWorker, "aegis.webhook")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize kafka consumer")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Run(ctx, webhookHandler(db, redis, &log)); err != nil {
			log.Error().Err(err).Msg("Relay service stopped with error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down Webhook Worker...")
	cancel()

	log.Info().Msg("Webhook Worker shutdown complete")
}
