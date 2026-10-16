package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerIntervalTask(t *testing.T) {
	s := New()

	var count int64
	taskID := s.Every(20*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt64(&count, 1)
		return nil
	})

	if taskID == "" {
		t.Fatalf("expected non-empty taskID")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	time.Sleep(65 * time.Millisecond)
	cancel()
	s.Stop()

	if val := atomic.LoadInt64(&count); val < 2 {
		t.Errorf("expected count >= 2, got: %d", val)
	}
}

func TestCronMatching(t *testing.T) {
	now := time.Date(2026, time.January, 15, 10, 30, 0, 0, time.UTC)

	if !matchCron("* * * * *", now) {
		t.Errorf("expected * * * * * to match")
	}
	if !matchCron("30 10 * * *", now) {
		t.Errorf("expected 30 10 * * * to match")
	}
	if !matchCron("*/15 * * * *", now) {
		t.Errorf("expected */15 * * * * to match")
	}
	if matchCron("45 * * * *", now) {
		t.Errorf("expected 45 * * * * NOT to match")
	}
}
