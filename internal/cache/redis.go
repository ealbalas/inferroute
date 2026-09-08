package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the go-redis client with domain helpers.
type Client struct {
	rdb *redis.Client
}

// New creates a Redis client from a URL (e.g. "redis://localhost:6379").
func New(url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Get retrieves a cached value. Returns ("", false, nil) on cache miss.
func (c *Client) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// Set stores a value with an expiry.
func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Del removes a key.
func (c *Client) Del(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

// IncrBy atomically increments a counter and sets a TTL on first creation.
// Used for sliding-window rate limiting.
func (c *Client) IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	pipe := c.rdb.Pipeline()
	incr := pipe.IncrBy(ctx, key, delta)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// CacheKey builds a deterministic cache key from a prompt and model name.
// The prompt itself is hashed so the key is always a fixed length.
func CacheKey(prompt, model string) string {
	sum := sha256.Sum256([]byte(prompt + "|" + model))
	return "cache:response:" + hex.EncodeToString(sum[:])
}

// RateLimitKey builds the Redis key for a user's rate limit window.
func RateLimitKey(userID string, window time.Time) string {
	return fmt.Sprintf("ratelimit:%s:%d", userID, window.Unix())
}

// Raw exposes the underlying redis.Client for advanced operations.
func (c *Client) Raw() *redis.Client { return c.rdb }
