package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file - ignore error if file doesn't exist
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Note: .env file not found or could not be loaded: %v\n", err)
	}
}

type Config struct {
	Primary       PrimaryConfig
	Database      DatabaseConfig
	Server        ServerConfig
	Redis         RedisConfig
	Observability *ObservabilityConfig
	Paystack      PaystackConfig
}

type PrimaryConfig struct {
	Env string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

type ServerConfig struct {
	Port               string
	ReadTimeout        int
	WriteTimeout       int
	IdleTimeout        int
	CORSAllowedOrigins []string
}

type RedisConfig struct {
	Address      string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LockTTL      time.Duration
	KeyPrefix    string
}

type ObservabilityConfig struct {
	ServiceName  string
	Environment  string
	Logging      LoggingConfig
	NewRelic     NewRelicConfig
	HealthChecks HealthChecksConfig
}

type LoggingConfig struct {
	Level              string
	Format             string
	SlowQueryThreshold time.Duration
}

type NewRelicConfig struct {
	LicenseKey                string
	AppLogForwardingEnabled   bool
	DistributedTracingEnabled bool
	DebugLogging              bool
}

type HealthChecksConfig struct {
	Enabled  bool
	Interval time.Duration
	Timeout  time.Duration
	Checks   []string
}

type PaystackConfig struct {
	SecretKey     string
	PublicKey     string
	WebhookSecret string
	BaseURL       string
}

// Helper functions for parsing env vars
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return fallback
}

func (c *ObservabilityConfig) GetLogLevel() string {
	if c.Logging.Level == "" {
		switch c.Environment {
		case "production":
			return "info"
		case "development":
			return "debug"
		default:
			return "info"
		}
	}
	return c.Logging.Level
}

func (c *ObservabilityConfig) IsProduction() bool {
	return c.Environment == "production"
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Primary: PrimaryConfig{
			Env: getEnv("AEGIS_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("AEGIS_DB_HOST", "localhost"),
			Port:            getEnvInt("AEGIS_DB_PORT", 5432),
			User:            getEnv("AEGIS_DB_USER", "aegis"),
			Password:        getEnv("AEGIS_DB_PASSWORD", ""),
			Name:            getEnv("AEGIS_DB_NAME", "aegis"),
			SSLMode:         getEnv("AEGIS_DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("AEGIS_DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("AEGIS_DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvInt("AEGIS_DB_CONN_MAX_LIFETIME", 300),
			ConnMaxIdleTime: getEnvInt("AEGIS_DB_CONN_MAX_IDLE_TIME", 60),
		},
		Server: ServerConfig{
			Port:               getEnv("AEGIS_SERVER_PORT", "8080"),
			ReadTimeout:        getEnvInt("AEGIS_SERVER_READ_TIMEOUT", 30),
			WriteTimeout:       getEnvInt("AEGIS_SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:        getEnvInt("AEGIS_SERVER_IDLE_TIMEOUT", 60),
			CORSAllowedOrigins: getEnvSlice("AEGIS_SERVER_CORS_ORIGINS", []string{"*"}),
		},
		Redis: RedisConfig{
			Address:      getEnv("AEGIS_REDIS_ADDRESS", "localhost:6379"),
			Password:     getEnv("AEGIS_REDIS_PASSWORD", ""),
			DB:           getEnvInt("AEGIS_REDIS_DB", 0),
			PoolSize:     getEnvInt("AEGIS_REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("AEGIS_REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:  getEnvDuration("AEGIS_REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("AEGIS_REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("AEGIS_REDIS_WRITE_TIMEOUT", 3*time.Second),
			LockTTL:      getEnvDuration("AEGIS_REDIS_LOCK_TTL", 30*time.Second),
			KeyPrefix:    getEnv("AEGIS_REDIS_KEY_PREFIX", "aegis:"),
		},
		Observability: &ObservabilityConfig{
			ServiceName: "Aegis",
			Environment: getEnv("AEGIS_ENV", "development"),
			Logging: LoggingConfig{
				Level:              getEnv("AEGIS_LOG_LEVEL", "debug"),
				Format:             getEnv("AEGIS_LOG_FORMAT", "console"),
				SlowQueryThreshold: getEnvDuration("AEGIS_LOG_SLOW_QUERY_THRESHOLD", 100*time.Millisecond),
			},
			NewRelic: NewRelicConfig{
				LicenseKey:                getEnv("AEGIS_NEWRELIC_LICENSE_KEY", ""),
				AppLogForwardingEnabled:   getEnvBool("AEGIS_NEWRELIC_LOG_FORWARDING", true),
				DistributedTracingEnabled: getEnvBool("AEGIS_NEWRELIC_DISTRIBUTED_TRACING", true),
				DebugLogging:              getEnvBool("AEGIS_NEWRELIC_DEBUG", false),
			},
			HealthChecks: HealthChecksConfig{
				Enabled:  getEnvBool("AEGIS_HEALTHCHECK_ENABLED", true),
				Interval: getEnvDuration("AEGIS_HEALTHCHECK_INTERVAL", 30*time.Second),
				Timeout:  getEnvDuration("AEGIS_HEALTHCHECK_TIMEOUT", 5*time.Second),
				Checks:   getEnvSlice("AEGIS_HEALTHCHECK_CHECKS", []string{"database", "redis"}),
			},
		},
		Paystack: PaystackConfig{
			SecretKey:     getEnv("AEGIS_PAYSTACK_SECRET_KEY", ""),
			PublicKey:     getEnv("AEGIS_PAYSTACK_PUBLIC_KEY", ""),
			WebhookSecret: getEnv("AEGIS_PAYSTACK_WEBHOOK_SECRET", ""),
			BaseURL:       getEnv("AEGIS_PAYSTACK_BASE_URL", "https://api.paystack.co"),
		},
	}

	// Validate required fields
	if cfg.Database.Host == "" {
		return nil, fmt.Errorf("AEGIS_DB_HOST is required")
	}
	if cfg.Database.Name == "" {
		return nil, fmt.Errorf("AEGIS_DB_NAME is required")
	}

	return cfg, nil
}
