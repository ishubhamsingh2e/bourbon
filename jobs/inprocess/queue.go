// Package inprocess provides an in-memory job queue with delays and retries.
package inprocess

import (
	"context"
	"sync"
	"time"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/jobs"
)

type queuedItem struct {
	job         jobs.Job
	opts        jobs.JobOptions
	availableAt time.Time
	attempts    int
}

// Queue implements jobs.Queue in-memory using a thread-safe slice.
type Queue struct {
	mu    sync.Mutex
	items map[string][]queuedItem
}

// New constructs an in-memory Queue.
func New() *Queue {
	return &Queue{
		items: make(map[string][]queuedItem),
	}
}

// Enqueue adds a job to the specified queue partition with options.
func (q *Queue) Enqueue(ctx context.Context, j jobs.Job, opts ...jobs.JobOption) error {
	options := jobs.DefaultJobOptions()
	for _, opt := range opts {
		opt(&options)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	item := queuedItem{
		job:         j,
		opts:        options,
		availableAt: time.Now().Add(options.Delay),
		attempts:    0,
	}

	q.items[options.Queue] = append(q.items[options.Queue], item)
	return nil
}

// Dequeue retrieves and removes the oldest available job from the specified queues.
func (q *Queue) Dequeue(ctx context.Context, queueNames ...string) (jobs.Job, error) {
	if len(queueNames) == 0 {
		queueNames = []string{"default"}
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()

	for _, queueName := range queueNames {
		items := q.items[queueName]
		for i, item := range items {
			if now.After(item.availableAt) || now.Equal(item.availableAt) {
				// Remove item from queue
				q.items[queueName] = append(items[:i], items[i+1:]...)
				return item.job, nil
			}
		}
	}

	return nil, jobs.ErrQueueEmpty
}

// Size returns the count of ready jobs in the specified queues.
func (q *Queue) Size(ctx context.Context, queueNames ...string) (int64, error) {
	if len(queueNames) == 0 {
		queueNames = []string{"default"}
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	var count int64

	for _, queueName := range queueNames {
		for _, item := range q.items[queueName] {
			if now.After(item.availableAt) || now.Equal(item.availableAt) {
				count++
			}
		}
	}

	return count, nil
}

// Provider implements core.Provider for in-process job queue integration.
type Provider struct {
	queue *Queue
}

// NewProvider returns a new in-process queue Provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Register registers jobs.Queue into the DI container as a singleton.
func (p *Provider) Register(c core.Container) error {
	p.queue = New()
	c.Singleton(func() jobs.Queue {
		return p.queue
	})
	return nil
}

// Boot verifies queue readiness.
func (p *Provider) Boot(ctx context.Context, app core.Application) error {
	var q jobs.Queue
	return app.Container().Resolve(&q)
}

// Shutdown releases queue resources.
func (p *Provider) Shutdown(ctx context.Context) error {
	return nil
}

var _ core.Provider = (*Provider)(nil)
