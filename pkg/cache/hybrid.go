package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// HybridCache combines local (L1) and remote (L2) caching
type HybridCache struct {
	local  Cache
	remote Cache
	group  singleflight.Group
	l1Inv  *L1Invalidator // L1 invalidator for cross-pod sync
}

// NewHybridCache creates a new HybridCache
func NewHybridCache(local Cache, remote Cache) *HybridCache {
	return &HybridCache{
		local:  local,
		remote: remote,
	}
}

func (c *HybridCache) Get(ctx context.Context, key string) (string, error) {
	// 1. Try L1
	val, err := c.local.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// 2. Try L2
	val, err = c.remote.Get(ctx, key)
	if err == nil {
		// Populate L1 (TTL? Default short or same as remotes?)
		// We don't know expiration here. HybridCache needs policy.
		// For simplicity, we assume L1 TTL is shorter or standard (e.g., 5 min).
		// We'll use a Safe default of 5 minutes for L1 Population on read.
		c.local.Set(ctx, key, val, 5*time.Minute)
		return val, nil
	}

	return "", err
}

func (c *HybridCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, _ := c.local.Exists(ctx, key)
	if exists {
		return true, nil
	}
	return c.remote.Exists(ctx, key)
}

func (c *HybridCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Set both
	// L1 expiration typically matches L2 or is capped.
	// We'll just pass expiration to both.
	if err := c.local.Set(ctx, key, value, expiration); err != nil {
		// Log?
	}
	return c.remote.Set(ctx, key, value, expiration)
}

func (c *HybridCache) Del(ctx context.Context, key string) error {
	c.local.Del(ctx, key)
	return c.remote.Del(ctx, key)
}

func (c *HybridCache) GetOrSet(ctx context.Context, key string, expiration time.Duration, fetch func() (string, error)) (string, error) {
	// 1. Try L1
	val, err := c.local.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// 2. Try L2
	val, err = c.remote.Get(ctx, key)
	if err == nil {
		c.local.Set(ctx, key, val, expiration) // Set L1
		return val, nil
	}

	// 3. Simple Miss or Error in both, use SingleFlight
	v, err, _ := c.group.Do(key, func() (interface{}, error) {
		// Fetch
		data, err := fetch()
		if err != nil {
			return "", err
		}

		// Set L2
		if err := c.remote.Set(ctx, key, data, expiration); err != nil {
			// Log
		}
		// Set L1
		c.local.Set(ctx, key, data, expiration)

		return data, nil
	})

	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// StartL1Invalidator starts the L1 cache invalidator that subscribes to Redis Pub/Sub
// for cross-pod cache synchronization. This should be called after creating HybridCache.
func (c *HybridCache) StartL1Invalidator(rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	c.l1Inv = NewL1Invalidator(c.local, rdb)
	return c.l1Inv.Start()
}

// StopL1Invalidator stops the L1 cache invalidator
func (c *HybridCache) StopL1Invalidator() error {
	if c.l1Inv != nil {
		return c.l1Inv.Stop()
	}
	return nil
}

func (c *HybridCache) Close() error {
	// Stop L1 invalidator first
	if c.l1Inv != nil {
		c.l1Inv.Stop()
	}
	c.local.Close()
	return c.remote.Close()
}

// Passthrough for complex Redis commands to Remote
func (c *HybridCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.remote.Incr(ctx, key)
}
func (c *HybridCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.remote.RPush(ctx, key, values...)
}
func (c *HybridCache) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return c.remote.BLPop(ctx, timeout, keys...)
}
func (c *HybridCache) LTrim(ctx context.Context, key string, start, stop int64) error {
	return c.remote.LTrim(ctx, key, start, stop)
}
func (c *HybridCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.remote.Eval(ctx, script, keys, args...)
}
func (c *HybridCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.remote.SAdd(ctx, key, members...)
}
func (c *HybridCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.remote.SIsMember(ctx, key, member)
}

func (c *HybridCache) Stats(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"local":  c.local.Stats(ctx),
		"remote": c.remote.Stats(ctx),
	}
}

// GetRedisClient attempts to return the underlying Redis client if available
func (c *HybridCache) GetRedisClient() *redis.Client {
	if rc, ok := c.remote.(*RedisCache); ok {
		return rc.Client()
	}
	return nil
}
