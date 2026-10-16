// Package jobs provides background job queues, worker pools, and execution contracts.
package jobs

import (
	"context"
	"time"
)

// Job defines the contract for background executable tasks.
type Job interface {
	Name() string
	Payload() []byte
	Execute(ctx context.Context) error
}

// JobOptions configures queue routing, retry policies, and execution delays.
type JobOptions struct {
	Queue      string
	MaxRetries int
	Backoff    time.Duration
	Delay      time.Duration
	Timeout    time.Duration
}

// DefaultJobOptions returns sensible background execution defaults.
func DefaultJobOptions() JobOptions {
	return JobOptions{
		Queue:      "default",
		MaxRetries: 3,
		Backoff:    time.Second,
		Delay:      0,
		Timeout:    30 * time.Second,
	}
}

// JobOption applies a configuration modification to JobOptions.
type JobOption func(*JobOptions)

// WithQueue specifies the target queue partition.
func WithQueue(name string) JobOption {
	return func(o *JobOptions) {
		if name != "" {
			o.Queue = name
		}
	}
}

// WithRetries configures the maximum retry attempts and backoff duration on failure.
func WithRetries(count int, backoff time.Duration) JobOption {
	return func(o *JobOptions) {
		o.MaxRetries = count
		o.Backoff = backoff
	}
}

// WithDelay postpones job execution until duration has elapsed.
func WithDelay(d time.Duration) JobOption {
	return func(o *JobOptions) {
		o.Delay = d
	}
}

// WithTimeout sets a maximum execution duration before canceling the context.
func WithTimeout(d time.Duration) JobOption {
	return func(o *JobOptions) {
		o.Timeout = d
	}
}

// BaseJob is a helper implementation of the Job interface.
type BaseJob struct {
	JobName    string
	JobPayload []byte
	Handler    func(ctx context.Context, payload []byte) error
}

// NewJob creates a BaseJob with a functional execution handler.
func NewJob(name string, payload []byte, handler func(ctx context.Context, payload []byte) error) BaseJob {
	return BaseJob{
		JobName:    name,
		JobPayload: payload,
		Handler:    handler,
	}
}

// Name returns the job identifier.
func (b BaseJob) Name() string {
	return b.JobName
}

// Payload returns the serialized job data.
func (b BaseJob) Payload() []byte {
	return b.JobPayload
}

// Execute invokes the underlying job handler function.
func (b BaseJob) Execute(ctx context.Context) error {
	if b.Handler != nil {
		return b.Handler(ctx, b.JobPayload)
	}
	return nil
}
