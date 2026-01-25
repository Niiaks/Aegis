package redis

import (
	"context"
	"fmt"
	"time"
)

// Lock represents a distributed lock
type Lock struct {
	client *Client
	key    string
	value  string
}

// AcquireLock attempts to acquire a distributed lock with a timeout
func (c *Client) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	prefixedKey := c.prefixKey("lock:" + key)
	value := fmt.Sprintf("%d", time.Now().UnixNano())

	ok, err := c.rdb.SetNX(ctx, prefixedKey, value, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !ok {
		return nil, fmt.Errorf("lock already held")
	}

	return &Lock{
		client: c,
		key:    prefixedKey,
		value:  value,
	}, nil
}

// Release releases the lock if it is still held by the owner
func (l *Lock) Release(ctx context.Context) error {
	// Simple script to ensure only the owner can release the lock
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
	`
	_, err := l.client.rdb.Eval(ctx, script, []string{l.key}, l.value).Result()
	return err
}
