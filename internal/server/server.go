package server

import (
	"fmt"

	"github.com/Niiaks/Aegis/internal/config"
	"github.com/Niiaks/Aegis/internal/database"
	loggerPkg "github.com/Niiaks/Aegis/internal/logger"
	"github.com/rs/zerolog"
)

type Server struct {
	Config        *config.Config
	Logger        *zerolog.Logger
	LoggerService *loggerPkg.LoggerService
	Db            *database.Database
}

func NewServer(cfg *config.Config, logger *zerolog.Logger, ls *loggerPkg.LoggerService, db *database.Database) (*Server, error) {
	db, err := database.New(cfg, logger, ls)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Server{
		Config:        cfg,
		Logger:        logger,
		LoggerService: ls,
		Db:            db,
	}, nil
}
