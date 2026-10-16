package jobs

import (
	"context"
	"errors"
)

var (
	// ErrQueueEmpty is returned when Dequeue is called on an empty queue.
	ErrQueueEmpty = errors.New("jobs: queue is empty")
)

// Queue defines the contract for background job persistence and retrieval.
type Queue interface {
	// Enqueue pushes a job into the queue with options.
	Enqueue(ctx context.Context, j Job, opts ...JobOption) error

	// Dequeue retrieves and removes the next ready job from the specified queues.
	Dequeue(ctx context.Context, queues ...string) (Job, error)

	// Size returns the count of ready jobs in the specified queues.
	Size(ctx context.Context, queues ...string) (int64, error)
}
