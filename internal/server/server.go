package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Niiaks/Aegis/internal/config"
	"github.com/Niiaks/Aegis/internal/database"
	loggerPkg "github.com/Niiaks/Aegis/internal/logger"
	"github.com/Niiaks/Aegis/internal/redis"
	"github.com/rs/zerolog"
)

type Server struct {
	Config        *config.Config
	Logger        *zerolog.Logger
	httpServer    *http.Server
	LoggerService *loggerPkg.LoggerService
	Db            *database.Database
	redis         *redis.Client
}

func NewServer(cfg *config.Config, logger *zerolog.Logger, ls *loggerPkg.LoggerService, db *database.Database, redis *redis.Client) (*Server, error) {
	return &Server{
		Config:        cfg,
		Logger:        logger,
		LoggerService: ls,
		Db:            db,
		redis:         redis,
	}, nil
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	s.httpServer = &http.Server{
		Addr:         ":" + s.Config.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.Server.IdleTimeout) * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return fmt.Errorf("HTTP server is not set up")
	}

	s.Logger.Info().Msgf("Starting server on port %s in %s", s.Config.Server.Port, s.Config.Primary.Env)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.Db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	if err := s.redis.Close(); err != nil {
		return fmt.Errorf("failed to close redis client: %w", err)
	}
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}
	return nil
}
