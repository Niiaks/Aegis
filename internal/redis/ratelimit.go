package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// RateLimitResult contains rate limit check result
type RateLimitResult struct {
	Allowed   bool
	Remaining int64
	ResetAt   time.Time
}

// CheckRateLimit implements a sliding window rate limiter
// key: unique identifier (e.g., "user:123", "ip:1.2.3.4")
// limit: max requests allowed
// window: time window for the limit
func (c *Client) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	prefixedKey := c.prefixKey("ratelimit:" + key)
	now := time.Now()
	windowStart := now.Add(-window).UnixMilli()

	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])
		
		-- Remove old entries outside the window
		redis.call("ZREMRANGEBYSCORE", key, "-inf", window_start)
		
		-- Count current requests in window
		local count = redis.call("ZCARD", key)
		
		if count < limit then
			-- Add this request
			redis.call("ZADD", key, now, now .. "-" .. math.random())
			redis.call("PEXPIRE", key, window_ms)
			return {1, limit - count - 1}
		else
			return {0, 0}
		end
	`)

	result, err := script.Run(ctx, c.rdb, []string{prefixedKey},
		now.UnixMilli(),
		windowStart,
		limit,
		window.Milliseconds(),
	).Slice()
	if err != nil {
		return nil, err
	}

	allowed := result[0].(int64) == 1
	remaining := result[1].(int64)

	return &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   now.Add(window),
	}, nil
}

// SimpleRateLimit is a simpler fixed window rate limiter
// Faster but less accurate at window boundaries
func (c *Client) SimpleRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	prefixedKey := c.prefixKey("ratelimit:" + key)

	// Increment counter
	count, err := c.rdb.Incr(ctx, prefixedKey).Result()
	if err != nil {
		return false, err
	}

	// Set expiry on first request
	if count == 1 {
		c.rdb.Expire(ctx, prefixedKey, window)
	}

	return count <= limit, nil
}
