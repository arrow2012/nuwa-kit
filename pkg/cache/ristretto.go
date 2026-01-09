package cache

import (
	"context"
	"errors"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/sync/singleflight"
)

// RistrettoCache implements cache.Cache using In-Process Memory (Ristretto)
type RistrettoCache struct {
	cache *ristretto.Cache[string, any]
	group singleflight.Group
}

// NewRistrettoCache creates a new RistrettoCache
func NewRistrettoCache(maxKeys int64) (*RistrettoCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: maxKeys * 10, // Recommended: 10x max keys
		MaxCost:     maxKeys,      // Max keys (assuming cost 1 per key)
		BufferItems: 64,           // Recommended: 64 per buffer
		Metrics:     true,         // Enable metrics for monitoring
	})
	if err != nil {
		return nil, err
	}
	return &RistrettoCache{cache: cache}, nil
}

func (c *RistrettoCache) Get(ctx context.Context, key string) (string, error) {
	val, found := c.cache.Get(key)
	if !found {
		return "", errors.New("redis: nil") // Simulate Redis Nil for consistency? Or custom error.
		// Using "redis: nil" compatibility logic to swap easily.
		// A better way is defining ErrCacheMiss in cache package.
	}
	return val.(string), nil
}

func (c *RistrettoCache) Exists(ctx context.Context, key string) (bool, error) {
	_, found := c.cache.Get(key)
	return found, nil
}

func (c *RistrettoCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Ristretto supports TTL
	c.cache.SetWithTTL(key, value, 1, expiration)
	return nil
}

func (c *RistrettoCache) Del(ctx context.Context, key string) error {
	c.cache.Del(key)
	return nil
}

func (c *RistrettoCache) Close() error {
	c.cache.Close()
	return nil
}

// GetOrSet for L1
func (c *RistrettoCache) GetOrSet(ctx context.Context, key string, expiration time.Duration, fetch func() (string, error)) (string, error) {
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	v, err, _ := c.group.Do(key, func() (interface{}, error) {
		data, err := fetch()
		if err != nil {
			return "", err
		}
		c.Set(ctx, key, data, expiration)
		return data, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// No-op for Redis specific methods (List, Set, Eval, etc.)
// If these are called on L1, they will error or panic.
// Ideally usage of L1 is mostly KV.
func (c *RistrettoCache) Incr(ctx context.Context, key string) (int64, error) {
	return 0, errors.New("not supported")
}
func (c *RistrettoCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return errors.New("not supported")
}
func (c *RistrettoCache) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return nil, errors.New("not supported")
}
func (c *RistrettoCache) LTrim(ctx context.Context, key string, start, stop int64) error {
	return errors.New("not supported")
}
func (c *RistrettoCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return nil, errors.New("not supported")
}
func (c *RistrettoCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return errors.New("not supported")
}
func (c *RistrettoCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return false, errors.New("not supported")
}

func (c *RistrettoCache) Stats(ctx context.Context) map[string]interface{} {
	stats := c.cache.Metrics
	return map[string]interface{}{
		"hits":         stats.Hits(),
		"misses":       stats.Misses(),
		"keys_added":   stats.KeysAdded(),
		"keys_evicted": stats.KeysEvicted(),
		"cost_added":   stats.CostAdded(),
		"cost_evicted": stats.CostEvicted(),
		"ratio":        stats.Ratio(),
	}
}
