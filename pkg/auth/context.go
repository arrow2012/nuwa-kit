package auth

import "context"

type contextKey string

const (
	userIDKey   contextKey = "userID"
	usernameKey contextKey = "username"
	tenantIDKey contextKey = "tenantID"
)

// WithUserID returns a new context with the given user ID
func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext returns the user ID from the context
func UserIDFromContext(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(userIDKey).(int)
	return v, ok
}

// WithUsername returns a new context with the given username
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

// UsernameFromContext returns the username from the context
func UsernameFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(usernameKey).(string)
	return v, ok
}

// WithTenantID returns a new context with the given tenant ID
func WithTenantID(ctx context.Context, tenantID int) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// TenantIDFromContext returns the tenant ID from the context
func TenantIDFromContext(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(tenantIDKey).(int)
	return v, ok
}
