package jobs

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Worker orchestrates concurrent job consumption from one or more queues.
type Worker struct {
	queue        Queue
	concurrency  int
	queues       []string
	pollInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewWorker constructs a Worker pool for processing queued jobs.
func NewWorker(q Queue, concurrency int, queues ...string) *Worker {
	if concurrency <= 0 {
		concurrency = 4
	}
	if len(queues) == 0 {
		queues = []string{"default"}
	}

	return &Worker{
		queue:        q,
		concurrency:  concurrency,
		queues:       queues,
		pollInterval: 25 * time.Millisecond,
	}
}

// Start launches worker goroutines and processes jobs until context cancellation.
func (w *Worker) Start(ctx context.Context) error {
	w.ctx, w.cancel = context.WithCancel(ctx)

	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.runWorkerLoop()
	}

	<-w.ctx.Done()
	w.wg.Wait()
	return nil
}

func (w *Worker) runWorkerLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			job, err := w.queue.Dequeue(w.ctx, w.queues...)
			if err != nil {
				if errors.Is(err, ErrQueueEmpty) {
					time.Sleep(w.pollInterval)
					continue
				}
				// Context cancelled or unrecoverable error
				return
			}

			if job != nil {
				_ = job.Execute(w.ctx)
			}
		}
	}
}

// Stop gracefully signals all workers to stop after finishing active jobs.
func (w *Worker) Stop() error {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	return nil
}
