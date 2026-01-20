package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Niiaks/Aegis/internal/config"
	"github.com/Niiaks/Aegis/internal/database"
	"github.com/Niiaks/Aegis/internal/logger"
	"github.com/Niiaks/Aegis/internal/redis"
	"github.com/Niiaks/Aegis/internal/router"
	"github.com/Niiaks/Aegis/internal/server"
	"github.com/Niiaks/Aegis/internal/user"
	"github.com/Niiaks/Aegis/internal/wallet"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	loggerService := logger.New(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	db, err := database.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}

	_, err = redis.New(&log, &cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize redis client")
	}
	srv, err := server.NewServer(cfg, &log, loggerService, db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}

	userRepo := user.NewUserRepository(db.Pool)
	walletRepo := wallet.NewWalletRepository(db.Pool)

	userService := user.NewUserService(userRepo)
	walletService := wallet.NewWalletService(walletRepo)

	userHandler := user.NewUserHandler(userService)
	walletHandler := wallet.NewWalletHandler(walletService)

	handlers := &router.Handlers{
		User:   userHandler,
		Wallet: walletHandler,
	}

	r := router.NewRouter(srv, handlers)

	srv.SetupHTTPServer(r)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	log.Info().Msg("server stopped")
}
