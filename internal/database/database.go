package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/Niiaks/Aegis/internal/config"
	loggerConfig "github.com/Niiaks/Aegis/internal/logger"
	pgxzero "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	nrpgx5 "github.com/newrelic/go-agent/v3/integrations/nrpgx5"
	"github.com/rs/zerolog"
)

type Database struct {
	Pool   *pgxpool.Pool
	logger *zerolog.Logger
}

// Multitracer allows chaining multiple tracers
type MultiTracer struct {
	tracers []pgx.QueryTracer
}

// TraceQueryStart calls TraceQueryStart on all tracers
func (mt *MultiTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, tracer := range mt.tracers {
		if t, ok := tracer.(interface {
			TraceQueryStart(context.Context, *pgx.Conn, pgx.TraceQueryStartData) context.Context
		}); ok {
			ctx = t.TraceQueryStart(ctx, conn, data)
		}
	}
	return ctx
}

// TraceQueryEnd calls TraceQueryEnd on all tracers
func (mt *MultiTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, tracer := range mt.tracers {
		if t, ok := tracer.(interface {
			TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData)
		}); ok {
			t.TraceQueryEnd(ctx, conn, data)
		}
	}
}

const DatabaseTimeout = 10

func New(cfg *config.Config, logger *zerolog.Logger, ls *loggerConfig.LoggerService) (*Database, error) {

	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// URL encode password.
	encodedPassword := url.QueryEscape(cfg.Database.Password)

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", cfg.Database.User, encodedPassword, hostPort, cfg.Database.Name, cfg.Database.SSLMode)

	pgxPoolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database dsn: %w", err)
	}

	// Add New Relic PostgreSQL config.
	if ls != nil && ls.GetApplication() != nil {
		pgxPoolConfig.ConnConfig.Tracer = nrpgx5.NewTracer()
	}

	if cfg.Primary.Env == "development" || cfg.Primary.Env == "staging" {
		globalLevel := logger.GetLevel()
		pgxLogger := loggerConfig.NewPgxLogger(globalLevel)

		if pgxPoolConfig.ConnConfig.Tracer != nil {
			// If New Relic tracer exists, create a multi-tracer
			localTracer := &tracelog.TraceLog{
				Logger:   pgxzero.NewLogger(pgxLogger),
				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLevel)),
			}
			pgxPoolConfig.ConnConfig.Tracer = &MultiTracer{
				tracers: []pgx.QueryTracer{pgxPoolConfig.ConnConfig.Tracer, localTracer},
			}
		} else {
			pgxPoolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
				Logger:   pgxzero.NewLogger(pgxLogger),
				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLevel)),
			}
		}
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	database := &Database{
		Pool:   pool,
		logger: logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), DatabaseTimeout*time.Second)
	defer cancel()
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("connected to the database")
	return database, nil
}

func (db *Database) Close() error {
	db.logger.Info().Msg("closing database connection pool")
	db.Pool.Close()
	return nil
}
