package inprocess

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/jobs"
)

func TestInProcessQueueAndWorker(t *testing.T) {
	q := New()
	ctx := context.Background()

	var executedCount int64

	job1 := jobs.NewJob("send_welcome", []byte("user_1"), func(ctx context.Context, payload []byte) error {
		atomic.AddInt64(&executedCount, 1)
		return nil
	})

	job2 := jobs.NewJob("send_welcome", []byte("user_2"), func(ctx context.Context, payload []byte) error {
		atomic.AddInt64(&executedCount, 1)
		return nil
	})

	// 1. Enqueue jobs
	_ = q.Enqueue(ctx, job1)
	_ = q.Enqueue(ctx, job2)

	size, _ := q.Size(ctx)
	if size != 2 {
		t.Errorf("queue size = %d, want 2", size)
	}

	// 2. Run worker
	worker := jobs.NewWorker(q, 2)
	wCtx, cancel := context.WithCancel(ctx)

	go func() {
		_ = worker.Start(wCtx)
	}()

	// Wait for worker to consume jobs
	time.Sleep(100 * time.Millisecond)
	cancel()
	_ = worker.Stop()

	if atomic.LoadInt64(&executedCount) != 2 {
		t.Errorf("executedCount = %d, want 2", executedCount)
	}

	finalSize, _ := q.Size(ctx)
	if finalSize != 0 {
		t.Errorf("final queue size = %d, want 0", finalSize)
	}
}

func TestInProcessQueueDelay(t *testing.T) {
	q := New()
	ctx := context.Background()

	delayedJob := jobs.NewJob("delayed_task", []byte("payload"), nil)
	_ = q.Enqueue(ctx, delayedJob, jobs.WithDelay(50*time.Millisecond))

	// Immediately should be empty because of delay
	_, err := q.Dequeue(ctx)
	if err == nil {
		t.Errorf("expected ErrQueueEmpty for delayed job before time, got nil")
	}

	// After delay, job should be available
	time.Sleep(60 * time.Millisecond)
	j, err := q.Dequeue(ctx)
	if err != nil || j == nil {
		t.Fatalf("expected job to be dequeued after delay, got: %v", err)
	}
}

func TestQueueProviderLifecycle(t *testing.T) {
	provider := NewProvider()
	application := app.New()

	if err := provider.Register(application.Container()); err != nil {
		t.Fatalf("failed registering queue provider: %v", err)
	}

	var q jobs.Queue
	if err := application.Container().Resolve(&q); err != nil {
		t.Fatalf("failed resolving jobs.Queue: %v", err)
	}

	ctx := context.Background()
	if err := provider.Boot(ctx, application); err != nil {
		t.Fatalf("failed booting queue provider: %v", err)
	}

	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("failed shutting down queue provider: %v", err)
	}
}
