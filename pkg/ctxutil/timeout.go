package ctxutil

import (
	"context"
	"time"
)

const (
	// DefaultQueryTimeout is the default timeout for database queries
	DefaultQueryTimeout = 5 * time.Second

	// LongQueryTimeout is for complex queries that may take longer
	LongQueryTimeout = 10 * time.Second
)

// WithQueryTimeout adds a timeout to the context for database queries
// Returns the new context and a cancel function that should be deferred
func WithQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, DefaultQueryTimeout)
}

// WithLongQueryTimeout adds a longer timeout for complex queries
func WithLongQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, LongQueryTimeout)
}

// WithCustomTimeout adds a custom timeout duration
func WithCustomTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
