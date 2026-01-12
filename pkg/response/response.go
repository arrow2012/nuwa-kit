package response

import (
	"net/http"

	"github.com/arrow2012/nuwa-kit/pkg/errors"
	"github.com/gin-gonic/gin"
)

// Response standard structure
type Response struct {
	Code      int         `json:"code"`                 // Business Code
	Message   string      `json:"message"`              // Message
	Data      interface{} `json:"data,omitempty"`       // Data payload
	RequestID string      `json:"request_id,omitempty"` // Request ID for tracing
}

// PageResult standard structure for pagination
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ContextKeyRequestID must match the key used in middleware
const ContextKeyRequestID = "requestID"

// Success sends a success response
func Success(c *gin.Context, data interface{}) {
	rid := c.GetString(ContextKeyRequestID)
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   "success",
		Data:      data,
		RequestID: rid,
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	rid := c.GetString(ContextKeyRequestID)

	var apiErr errors.ErrorCode
	if e, ok := err.(errors.ErrorCode); ok {
		apiErr = e
	} else {
		// Default to Internal Server Error if unknown error type
		// But preserve the original message if safe, or general message
		// For security, maybe hide internal details in production
		// For now, let's wrap it as Internal Error
		apiErr = errors.ErrInternalServer
		// Optional: Log the real error here if not already logged
	}

	c.JSON(apiErr.HTTPStatus(), Response{
		Code:      apiErr.BusinessCode(),
		Message:   apiErr.Message(),
		RequestID: rid,
	})
}
