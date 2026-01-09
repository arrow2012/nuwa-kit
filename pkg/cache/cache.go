package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Cache interface defines the methods for caching
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) error
	BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error)
	LTrim(ctx context.Context, key string, start, stop int64) error
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	// GetOrSet retrieves the value from cache or executes the fetch function, catching stampedes
	GetOrSet(ctx context.Context, key string, expiration time.Duration, fetch func() (string, error)) (string, error)
	// Stats returns cache statistics
	Stats(ctx context.Context) map[string]interface{}
	Close() error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
	group  singleflight.Group
}

// NewRedisCache creates a new RedisCache
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Client returns the underlying redis client
func (c *RedisCache) Client() *redis.Client {
	return c.client
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.RPush(ctx, key, values...).Err()
}

func (c *RedisCache) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return c.client.BLPop(ctx, timeout, keys...).Result()
}

func (c *RedisCache) LTrim(ctx context.Context, key string, start, stop int64) error {
	return c.client.LTrim(ctx, key, start, stop).Err()
}

func (c *RedisCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

func (c *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

func (c *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) GetOrSet(ctx context.Context, key string, expiration time.Duration, fetch func() (string, error)) (string, error) {
	// 1. Try Cache
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}
	if err != redis.Nil {
		return "", err
	}

	// 2. Cache Miss: Use SingleFlight
	v, err, _ := c.group.Do(key, func() (interface{}, error) {
		// Fetch Data
		data, err := fetch()
		if err != nil {
			return "", err
		}

		// Set Cache
		if err := c.Set(ctx, key, data, expiration); err != nil {
			// Log error but continue
		}
		return data, nil
	})

	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (c *RedisCache) Stats(ctx context.Context) map[string]interface{} {
	stats := c.client.PoolStats()
	size, _ := c.client.DBSize(ctx).Result()

	return map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
		"items":       size,
		"byte_size":   int64(0), // Placeholder, requires parsing INFO memory
	}
}
