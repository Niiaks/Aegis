package redis

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrLockNotAcquired = errors.New("failed to acquire lock")
	ErrLockNotHeld     = errors.New("lock not held by this owner")
)

// Lock represents a distributed lock
type Lock struct {
	client *Client
	key    string
	owner  string
	ttl    time.Duration
}

// AcquireLock attempts to acquire a distributed lock
// Returns a Lock if successful, ErrLockNotAcquired if lock is held by another
func (c *Client) AcquireLock(ctx context.Context, resource string, ttl time.Duration) (*Lock, error) {
	key := c.prefixKey("lock:" + resource)
	owner := uuid.New().String()

	set, err := c.rdb.SetNX(ctx, key, owner, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !set {
		return nil, ErrLockNotAcquired
	}

	return &Lock{
		client: c,
		key:    key,
		owner:  owner,
		ttl:    ttl,
	}, nil
}

func (c *Client) TryAcquireLock(ctx context.Context, resource string, ttl time.Duration, maxRetries int, retryDelay time.Duration) (*Lock, error) {
	for i := 0; i <= maxRetries; i++ {
		lock, err := c.AcquireLock(ctx, resource, ttl)
		if err == nil {
			return lock, nil
		}
		if !errors.Is(err, ErrLockNotAcquired) {
			return nil, err
		}

		if i < maxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}
	return nil, ErrLockNotAcquired
}

// Release releases the lock (only if we own it)
func (l *Lock) Release(ctx context.Context) error {
	// Lua script to ensure we only delete if we own the lock
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.client.rdb, []string{l.key}, l.owner).Int()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// Extend extends the lock TTL (only if we own it)
func (l *Lock) Extend(ctx context.Context, ttl time.Duration) error {
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.client.rdb, []string{l.key}, l.owner, ttl.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrLockNotHeld
	}
	l.ttl = ttl
	return nil
}

// WithLock is a helper that acquires a lock, runs the function, and releases
func (c *Client) WithLock(ctx context.Context, resource string, ttl time.Duration, fn func() error) error {
	lock, err := c.AcquireLock(ctx, resource, ttl)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	return fn()
}
