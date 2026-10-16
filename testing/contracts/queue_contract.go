package contracts

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/jobs"
)

// VerifyJobQueueContract executes a standard compliance test suite against any jobs.Queue implementation.
func VerifyJobQueueContract(t *testing.T, queue jobs.Queue) {
	ctx := context.Background()

	t.Run("Enqueue_and_Dequeue", func(t *testing.T) {
		job := jobs.NewJob("contract_job", []byte("payload_123"), nil)

		err := queue.Enqueue(ctx, job)
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}

		size, err := queue.Size(ctx)
		if err != nil || size < 1 {
			t.Fatalf("expected queue size >= 1, got %d, err: %v", size, err)
		}

		dequeued, err := queue.Dequeue(ctx)
		if err != nil {
			t.Fatalf("Dequeue failed: %v", err)
		}
		if dequeued.Name() != "contract_job" {
			t.Errorf("dequeued.Name() = %v, want contract_job", dequeued.Name())
		}
	})
}
