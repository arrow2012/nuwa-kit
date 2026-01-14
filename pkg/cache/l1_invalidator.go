package cache

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// L1InvalidateChannel is the Redis Pub/Sub channel for L1 cache invalidation
	L1InvalidateChannel = "cache:l1:invalidate"
)

// L1Invalidator subscribes to Redis Pub/Sub and clears local L1 cache
// when invalidation messages are received. This ensures L1 cache consistency
// across multiple pods/instances sharing the same Redis.
type L1Invalidator struct {
	l1        Cache
	rdb       *redis.Client
	pubsub    *redis.PubSub
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.Mutex
}

// NewL1Invalidator creates a new L1Invalidator
func NewL1Invalidator(l1 Cache, rdb *redis.Client) *L1Invalidator {
	ctx, cancel := context.WithCancel(context.Background())
	return &L1Invalidator{
		l1:     l1,
		rdb:    rdb,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins listening for invalidation messages from Redis Pub/Sub.
// This should be called once during application startup.
func (i *L1Invalidator) Start() error {
	i.mu.Lock()
	if i.isRunning {
		i.mu.Unlock()
		return nil
	}
	i.isRunning = true
	i.mu.Unlock()

	// Subscribe to the invalidation channel
	i.pubsub = i.rdb.Subscribe(i.ctx, L1InvalidateChannel)

	// Wait for subscription confirmation
	_, err := i.pubsub.Receive(i.ctx)
	if err != nil {
		return err
	}

	i.wg.Add(1)
	go i.listen()

	return nil
}

// listen processes incoming invalidation messages
func (i *L1Invalidator) listen() {
	defer i.wg.Done()

	ch := i.pubsub.Channel()
	for {
		select {
		case <-i.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			i.handleMessage(msg)
		}
	}
}

// handleMessage processes a single invalidation message
func (i *L1Invalidator) handleMessage(msg *redis.Message) {
	if msg == nil {
		return
	}

	key := msg.Payload
	if key == "" {
		return
	}

	// Handle wildcard pattern (e.g., "iam:result:123:*")
	if strings.HasSuffix(key, ":*") {
		// For wildcard patterns, we need to clear all matching keys
		// Since Ristretto doesn't support pattern matching, we'll log this
		// and rely on the version-based invalidation for full cleanup
		// Wildcard patterns are handled by version-based invalidation
		return
	}

	// Delete the specific key from L1
	_ = i.l1.Del(context.Background(), key)
}

// Stop gracefully shuts down the invalidator
func (i *L1Invalidator) Stop() error {
	i.mu.Lock()
	if !i.isRunning {
		i.mu.Unlock()
		return nil
	}
	i.isRunning = false
	i.mu.Unlock()

	i.cancel()

	if i.pubsub != nil {
		_ = i.pubsub.Close()
	}

	// Wait for listener to stop with timeout
	done := make(chan struct{})
	go func() {
		i.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Stopped gracefully
	case <-time.After(5 * time.Second):
		// Timeout, force shutdown
	}

	return nil
}

// PublishInvalidation broadcasts a key invalidation to all pods via Redis Pub/Sub.
// This should be called whenever a cached value is updated or deleted.
func PublishInvalidation(ctx context.Context, rdb *redis.Client, key string) error {
	if rdb == nil {
		return nil
	}
	return rdb.Publish(ctx, L1InvalidateChannel, key).Err()
}

// PublishInvalidationMulti broadcasts multiple key invalidations
func PublishInvalidationMulti(ctx context.Context, rdb *redis.Client, keys ...string) error {
	if rdb == nil || len(keys) == 0 {
		return nil
	}

	pipe := rdb.Pipeline()
	for _, key := range keys {
		pipe.Publish(ctx, L1InvalidateChannel, key)
	}
	_, err := pipe.Exec(ctx)
	return err
}
