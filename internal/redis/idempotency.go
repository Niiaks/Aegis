package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyExists   = errors.New("idempotency key already exists")
	ErrKeyNotFound = errors.New("idempotency key not found")
)

// IdempotencyResult stores the result of an idempotent operation
type IdempotencyResult struct {
	Status   string // "pending", "completed", "failed"
	Response []byte // Cached response if completed
}

// SetIdempotencyKey sets a key if it doesn't exist (for idempotency check)
// Returns ErrKeyExists if key already exists
func (c *Client) SetIdempotencyKey(ctx context.Context, key string, ttl time.Duration) error {
	prefixedKey := c.prefixKey("idempotency:" + key)

	set, err := c.rdb.SetNX(ctx, prefixedKey, "pending", ttl).Result()
	if err != nil {
		return err
	}
	if !set {
		return ErrKeyExists
	}
	return nil
}

// GetIdempotencyKey retrieves an existing idempotency key's value
func (c *Client) GetIdempotencyKey(ctx context.Context, key string) (string, error) {
	prefixedKey := c.prefixKey("idempotency:" + key)

	val, err := c.rdb.Get(ctx, prefixedKey).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	}
	return val, err
}

// MarkIdempotencyComplete marks an idempotent operation as completed with response
func (c *Client) MarkIdempotencyComplete(ctx context.Context, key string, response []byte, ttl time.Duration) error {
	prefixedKey := c.prefixKey("idempotency:" + key)

	return c.rdb.Set(ctx, prefixedKey, response, ttl).Err()
}

func (c *Client) MarkIdempotencyFailed(ctx context.Context, key string) error {
	prefixedKey := c.prefixKey("idempotency:" + key)

	return c.rdb.Del(ctx, prefixedKey).Err()
}

// CheckAndSetIdempotency is a helper that checks if operation was already done
func (c *Client) CheckAndSetIdempotency(ctx context.Context, key string, ttl time.Duration) ([]byte, error) {
	prefixedKey := c.prefixKey("idempotency:" + key)

	set, err := c.rdb.SetNX(ctx, prefixedKey, "pending", ttl).Result()
	if err != nil {
		return nil, err
	}

	if set {
		return nil, nil
	}

	val, err := c.rdb.Get(ctx, prefixedKey).Result()
	if err != nil {
		return nil, err
	}

	if val == "pending" {
		return nil, ErrKeyExists
	}

	return []byte(val), nil
}
