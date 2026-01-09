package event

import (
	"context"
	"time"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Handler is a function that processes an event
type Handler func(ctx context.Context, event Event) error

// Bus defines the interface for an event bus
type Bus interface {
	// Publish publishes an event to a topic
	Publish(ctx context.Context, topic string, payload map[string]interface{}, metadata map[string]string) error

	// Subscribe subscribes to a topic with a consumer group
	Subscribe(ctx context.Context, topic string, group string, handler Handler) error

	// Close closes the bus connection
	Close() error
}
