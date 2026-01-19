package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Niiaks/Aegis/internal/config"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/nrzerolog"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type LoggerService struct {
	nrApp *newrelic.Application
}

func New(c *config.ObservabilityConfig) *LoggerService {
	service := &LoggerService{}

	if c.NewRelic.LicenseKey == "" {
		return service
	}

	var configurations []newrelic.ConfigOption

	configurations = append(configurations,
		newrelic.ConfigAppName(c.ServiceName),
		newrelic.ConfigLicense(c.NewRelic.LicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(c.NewRelic.AppLogForwardingEnabled),
		newrelic.ConfigDistributedTracerEnabled(c.NewRelic.DistributedTracingEnabled),
	)

	if c.NewRelic.DebugLogging {
		configurations = append(configurations, newrelic.ConfigDebugLogger(os.Stdout))
	}

	app, err := newrelic.NewApplication(configurations...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize New Relic: %v\n", err)
		return service
	}

	service.nrApp = app
	return service
}

// GetApplication returns the New Relic application instance
func (ls *LoggerService) GetApplication() *newrelic.Application {
	return ls.nrApp
}

// NewLoggerWithService creates a logger with full config and logger service
func NewLoggerWithService(cfg *config.ObservabilityConfig, loggerService *LoggerService) zerolog.Logger {
	var logLevel zerolog.Level
	level := cfg.GetLogLevel()

	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	// Let each logger have its own level
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var baseLogger zerolog.Logger

	// Setup base logger based on environment
	if cfg.IsProduction() && cfg.Logging.Format == "json" {
		// In production, write JSON to stdout
		baseLogger = zerolog.New(os.Stdout)
	} else {
		// Development mode - use console writer
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		baseLogger = zerolog.New(consoleWriter)
	}

	// Add New Relic hook for log forwarding in production
	if cfg.IsProduction() && loggerService != nil && loggerService.nrApp != nil {
		nrHook := nrzerolog.NewRelicHook{
			App: loggerService.nrApp,
		}
		baseLogger = baseLogger.Hook(nrHook)
	}

	logger := baseLogger.
		Level(logLevel).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Logger()

	// Include stack traces for errors in development
	if !cfg.IsProduction() {
		logger = logger.With().Stack().Logger()
	}

	return logger
}

// WithTraceContext adds New Relic transaction context to logger
func WithTraceContext(logger zerolog.Logger, txn *newrelic.Transaction) zerolog.Logger {
	if txn == nil {
		return logger
	}

	// Get trace metadata from transaction
	metadata := txn.GetTraceMetadata()

	return logger.With().
		Str("trace.id", metadata.TraceID).
		Str("span.id", metadata.SpanID).
		Logger()
}

// NewPgxLogger creates a database logger
func NewPgxLogger(level zerolog.Level) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatFieldValue: func(i any) string {
			switch v := i.(type) {
			case string:
				// Clean and format SQL for better readability
				if len(v) > 200 {
					// Truncate very long SQL statements
					return v[:200] + "..."
				}
				return v
			case []byte:
				var obj interface{}
				if err := json.Unmarshal(v, &obj); err == nil {
					pretty, _ := json.MarshalIndent(obj, "", "    ")
					return "\n" + string(pretty)
				}
				return string(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}

	return zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("component", "database").
		Logger()
}

// GetPgxTraceLogLevel converts zerolog level to pgx tracelog level
func GetPgxTraceLogLevel(level zerolog.Level) int {
	switch level {
	case zerolog.DebugLevel:
		return 6
	case zerolog.InfoLevel:
		return 4
	case zerolog.WarnLevel:
		return 3
	case zerolog.ErrorLevel:
		return 2
	default:
		return 0
	}
}

func (ls *LoggerService) Shutdown() {
	if ls.nrApp != nil {
		ls.nrApp.Shutdown(10 * time.Second)
	}
}
