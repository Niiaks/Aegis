package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/Niiaks/Aegis/internal/config"
)

type Client struct {
	rdb       *redis.Client
	keyPrefix string
	log       *zerolog.Logger
}

func New(log *zerolog.Logger, cfg *config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info().Msg("Connected to Redis successfully")

	return &Client{
		rdb:       rdb,
		keyPrefix: cfg.KeyPrefix,
		log:       log,
	}, nil
}

func (c *Client) prefixKey(key string) string {
	return c.keyPrefix + key
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	c.log.Info().Msg("Closing Redis client connection")
	return c.rdb.Close()
}
