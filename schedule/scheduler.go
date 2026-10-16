// Package schedule provides cron-style and interval-based task scheduling.
package schedule

import (
	"context"
	"time"
)

// TaskFunc defines the executable task function signature.
type TaskFunc func(ctx context.Context) error

// Scheduler defines the contract for registering and executing recurring tasks.
type Scheduler interface {
	// Every registers a task to run periodically at the specified duration interval.
	Every(interval time.Duration, task TaskFunc) string

	// Cron registers a task using standard cron expression syntax (5 fields: minute hour dom month dow).
	Cron(expression string, task TaskFunc) (string, error)

	// Cancel removes a scheduled task by its identifier.
	Cancel(taskID string)

	// Start begins task execution loops until context is cancelled.
	Start(ctx context.Context)

	// Stop terminates the scheduler and all active recurring tasks.
	Stop()
}
