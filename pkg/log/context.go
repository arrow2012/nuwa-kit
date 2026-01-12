package log

import (
	"context"

	"github.com/arrow2012/nuwa-kit/pkg/auth"
	"go.uber.org/zap"
)

// C returns a logger with context fields (trace_id, user_id, etc.)
func C(ctx context.Context) *zap.Logger {
	logger := L()
	if ctx == nil {
		return logger
	}

	var fields []zap.Field

	// TraceID
	if requestID := ctx.Value("request_id"); requestID != nil {
		if s, ok := requestID.(string); ok {
			fields = append(fields, zap.String("request_id", s))
		}
	}

	// UserID
	if userID, exists := auth.UserIDFromContext(ctx); exists {
		fields = append(fields, zap.Int("user_id", userID))
	}

	// Username
	if username, exists := auth.UsernameFromContext(ctx); exists {
		fields = append(fields, zap.String("username", username))
	}

	// TenantID
	if tenantID, exists := auth.TenantIDFromContext(ctx); exists {
		fields = append(fields, zap.Int("tenant_id", tenantID))
	}

	if len(fields) > 0 {
		return logger.With(fields...)
	}

	return logger
}

// CInfo logs a message with context fields
func CInfo(ctx context.Context, msg string, fields ...zap.Field) {
	C(ctx).Info(msg, fields...)
}

// CError logs a message with context fields
func CError(ctx context.Context, msg string, fields ...zap.Field) {
	C(ctx).Error(msg, fields...)
}

// CWarn logs a message with context fields
func CWarn(ctx context.Context, msg string, fields ...zap.Field) {
	C(ctx).Warn(msg, fields...)
}

// CDebug logs a message with context fields
func CDebug(ctx context.Context, msg string, fields ...zap.Field) {
	C(ctx).Debug(msg, fields...)
}
