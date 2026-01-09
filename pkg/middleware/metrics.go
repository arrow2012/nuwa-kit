package middleware

import (
	"strconv"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/metric"
	"github.com/gin-gonic/gin"
)

// Metrics records HTTP request metrics (count and duration).
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // Use FullPath to avoid high cardinality (e.g., /users/:id instead of /users/123)

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// If path is empty (404), use RequestURI but maybe generalize?
		if path == "" {
			path = "unknown"
		}

		metric.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metric.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
