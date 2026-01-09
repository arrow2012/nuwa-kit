package event

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/json"
	"github.com/arrow2012/nuwa-kit/pkg/log"
	"github.com/arrow2012/nuwa-kit/pkg/metric"
	"github.com/redis/go-redis/v9"
)

// RedisBus implements Bus using Redis Streams
type RedisBus struct {
	client redis.UniversalClient
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRedisBus creates a new RedisBus
func NewRedisBus(client redis.UniversalClient) *RedisBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisBus{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Publish publishes an event to a topic (Redis Stream)
func (b *RedisBus) Publish(ctx context.Context, topic string, payload map[string]interface{}, metadata map[string]string) error {
	event := Event{
		Topic:     topic,
		Payload:   payload,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Using MaxLenApprox to prevent unbounded growth (e.g., keep last 10000 events)
	err = b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{"event": data},
	}).Err()

	if err == nil {
		metric.EventBusPublished.WithLabelValues(topic).Inc()
	}

	return err
}

// Subscribe subscribes to a topic with a consumer group
func (b *RedisBus) Subscribe(ctx context.Context, topic string, group string, handler Handler) error {
	// 1. Create Consumer Group (idempotent ignore "BUSYGROUP")
	err := b.client.XGroupCreateMkStream(ctx, topic, group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}

	// 2. Start Consumer Loop
	b.wg.Add(1)
	go b.consumerLoop(ctx, topic, group, handler)

	return nil
}

func (b *RedisBus) consumerLoop(ctx context.Context, topic, group string, handler Handler) {
	defer b.wg.Done()

	// Unique consumer name (e.g., hostname + uuid)
	// For simplicity, using "consumer-1" but in prod should be unique per instance
	consumer := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	// We need to respect BOTH Bus context (b.ctx) and Subscription context (ctx)
	// If either is cancelled, we stop.
	// Since XReadGroup takes one context, passing 'ctx' (subscription context) is usually correct
	// because if b.ctx (global) is cancelled, we assume the caller will also likely cancel sub contexts,
	// OR we can just check b.ctx in select.
	// BETTER: If b.ctx is cancelled, we want to stop too.
	// So we select on both.

	for {
		select {
		case <-b.ctx.Done(): // Bus is closing
			return
		case <-ctx.Done(): // Subscription is cancelled (Worker stopped)
			return
		default:
			// Read from stream using passed ctx for cancellation
			streams, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{topic, ">"},
				Count:    10,
				Block:    500 * time.Millisecond,
			}).Result()

			if err != nil {
				if errors.Is(err, context.Canceled) {
					// Sub context cancelled
					return
				}
				// Also check if Bus context is done, just in case
				select {
				case <-b.ctx.Done():
					return
				default:
				}

				if err != redis.Nil {
					log.Errorf("RedisBus: Read error for topic %s: %v", topic, err)
					// Backoff
					select {
					case <-b.ctx.Done():
						return
					case <-ctx.Done():
						return
					case <-time.After(1 * time.Second):
					}
				}
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					var event Event
					eventData, ok := msg.Values["event"].(string)
					if !ok {
						log.Errorf("RedisBus: Invalid message format in topic %s", topic)
						metric.EventBusConsumed.WithLabelValues(topic, "invalid_format").Inc()
						b.client.XAck(context.Background(), topic, group, msg.ID)
						continue
					}

					if err := json.Unmarshal([]byte(eventData), &event); err != nil {
						log.Errorf("RedisBus: Failed to unmarshal event: %v", err)
						metric.EventBusConsumed.WithLabelValues(topic, "unmarshal_error").Inc()
						b.client.XAck(context.Background(), topic, group, msg.ID)
						continue
					}

					event.ID = msg.ID

					// Handle Event
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Errorf("panic: %v\n\n%s", r, string(debug.Stack()))
								metric.EventBusConsumed.WithLabelValues(topic, "panic").Inc()
							}
						}()
						if err := handler(ctx, event); err != nil {
							log.Errorf("RedisBus: Handler failed for topic %s: %v", topic, err)
							metric.EventBusConsumed.WithLabelValues(topic, "handler_failed").Inc()
						} else {
							metric.EventBusConsumed.WithLabelValues(topic, "success").Inc()
						}
					}()

					// Ack
					b.client.XAck(context.Background(), topic, group, msg.ID)
				}
			}
		}
	}
}

// Close closes the bus
func (b *RedisBus) Close() error {
	b.cancel() // Cancel context immediately to interrupt XReadGroup
	// log.Info("RedisBus: Waiting for consumers to stop...")
	b.wg.Wait()
	return nil
}
