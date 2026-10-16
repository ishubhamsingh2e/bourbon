package inprocess

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/events"
)

func TestInProcessBusPubSub(t *testing.T) {
	bus := New(100)
	defer bus.Close()

	var syncCount int64
	var asyncCount int64
	var wildcardCount int64

	bus.Subscribe("user.created", func(ctx context.Context, e events.Event) error {
		atomic.AddInt64(&syncCount, 1)
		return nil
	})

	bus.SubscribeAsync("user.created", func(ctx context.Context, e events.Event) error {
		atomic.AddInt64(&asyncCount, 1)
		return nil
	})

	bus.Subscribe("*", func(ctx context.Context, e events.Event) error {
		atomic.AddInt64(&wildcardCount, 1)
		return nil
	})

	event := events.NewEvent("user.created", map[string]any{"id": "u123"})
	ctx := context.Background()

	// Publish should invoke sync immediately and queue async
	err := bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if atomic.LoadInt64(&syncCount) != 1 {
		t.Errorf("syncCount = %d, want 1", syncCount)
	}
	if atomic.LoadInt64(&wildcardCount) != 1 {
		t.Errorf("wildcardCount = %d, want 1", wildcardCount)
	}

	// Wait for async worker
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt64(&asyncCount) != 1 {
		t.Errorf("asyncCount = %d, want 1", asyncCount)
	}
}

func TestBusProviderLifecycle(t *testing.T) {
	provider := NewProvider()
	application := app.New()

	if err := provider.Register(application.Container()); err != nil {
		t.Fatalf("failed registering bus provider: %v", err)
	}

	var b events.Bus
	if err := application.Container().Resolve(&b); err != nil {
		t.Fatalf("failed resolving events.Bus: %v", err)
	}

	ctx := context.Background()
	if err := provider.Boot(ctx, application); err != nil {
		t.Fatalf("failed booting bus provider: %v", err)
	}

	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("failed shutting down bus provider: %v", err)
	}
}
