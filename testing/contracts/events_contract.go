package contracts

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/events"
)

// VerifyEventBusContract executes a standard compliance test suite against any events.Bus implementation.
func VerifyEventBusContract(t *testing.T, bus events.Bus) {
	ctx := context.Background()

	t.Run("Dispatch_Sync", func(t *testing.T) {
		var received int64
		unsub := bus.Subscribe("contract.test", func(ctx context.Context, e events.Event) error {
			atomic.AddInt64(&received, 1)
			return nil
		})
		defer unsub()

		event := events.NewEvent("contract.test", "data")
		if err := bus.Dispatch(ctx, event); err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}

		if atomic.LoadInt64(&received) != 1 {
			t.Errorf("expected received == 1, got %d", received)
		}
	})

	t.Run("Publish_Async", func(t *testing.T) {
		var received int64
		unsub := bus.SubscribeAsync("contract.async", func(ctx context.Context, e events.Event) error {
			atomic.AddInt64(&received, 1)
			return nil
		})
		defer unsub()

		event := events.NewEvent("contract.async", "data")
		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf("Publish failed: %v", err)
		}

		time.Sleep(50 * time.Millisecond)
		if atomic.LoadInt64(&received) != 1 {
			t.Errorf("expected async received == 1, got %d", received)
		}
	})
}
