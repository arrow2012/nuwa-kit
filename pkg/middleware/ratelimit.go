package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// IPRateLimitMiddleware limits requests based on Client IP
// limit: max requests allowed per window
// window: time window duration
func IPRateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	limiter := ratelimit.NewLimiter(rdb)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:ip:%s", ip)

		res, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			// If Redis fails, fail open (allow request) or log error
			// For high availability, fail open is usually preferred
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", int(res.ResetAfter/time.Second)))

		if res.Allowed == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": int(res.RetryAfter / time.Second),
			})
			return
		}

		c.Next()
	}
}
