package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderXRequestID = "X-Request-ID"
	ContextRequestID = "request_id"
)

// RequestID adds a unique ID to every request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if ID is already in header
		rid := c.GetHeader(HeaderXRequestID)
		if rid == "" {
			rid = uuid.New().String()
		}

		// Set in Context
		c.Set(ContextRequestID, rid)

		// Set request header (crucial for downstream handlers like grpc-gateway)
		c.Request.Header.Set(HeaderXRequestID, rid)

		// Set response header
		c.Header(HeaderXRequestID, rid)

		c.Next()
	}
}
