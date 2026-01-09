package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/arrow2012/nuwa-kit/pkg/log"
	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				stack := string(debug.Stack())
				log.Errorf("panic recovered: %v\n%s", err, stack)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}
