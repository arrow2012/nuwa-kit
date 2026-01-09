package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// Limiter warps redis_rate.Limiter
type Limiter struct {
	limiter *redis_rate.Limiter
}

// NewLimiter creates a new rate limiter backed by Redis
func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{
		limiter: redis_rate.NewLimiter(rdb),
	}
}

// Allow checks if the request is allowed
// limit: max requests allowd
// window: time window
// Example: Allow("ip:127.0.0.1", 100, time.Minute) means 100 reqs per minute
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
	return l.limiter.Allow(ctx, key, redis_rate.Limit{
		Rate:   limit,
		Period: window,
		Burst:  limit,
	})
}
