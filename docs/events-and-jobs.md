# Events, Background Jobs & Real-Time Hub in Bourbon

Bourbon includes built-in asynchronous processing subsystems: an in-process event bus, background job queues with worker pools and retry backoff, a cron scheduler, and real-time Server-Sent Events (SSE).

---

## 1. In-Process Event Bus

The event bus supports synchronous dispatching and decoupled asynchronous event delivery:

```go
package events_example

import (
	"context"
	"log"

	"github.com/ishubhamsingh2e/bourbon/events"
	"github.com/ishubhamsingh2e/bourbon/events/inprocess"
)

func SetupEvents(ctx context.Context) {
	bus := inprocess.New(128)
	defer bus.Close()

	// 1. Subscribe listener to topic
	bus.Subscribe("user.registered", func(ctx context.Context, e events.Event) error {
		log.Printf("Received event: %s with payload: %s", e.Name(), string(e.Payload()))
		return nil
	})

	// 2. Synchronous dispatch (blocks until all subscribers finish)
	event := events.New("user.registered", []byte(`{"user_id":"u_123","email":"user@example.com"}`))
	if err := bus.Dispatch(ctx, event); err != nil {
		log.Printf("dispatch error: %v", err)
	}

	// 3. Asynchronous publish (non-blocking)
	_ = bus.Publish(ctx, event)
}
```

---

## 2. Background Job Queues & Worker Pool

Bourbon's job subsystem handles background task processing, concurrency limits, delayed execution, and retry exponential backoffs without requiring Redis for basic setups:

```go
package jobs_example

import (
	"context"
	"log"
	"time"

	"github.com/ishubhamsingh2e/bourbon/jobs"
	"github.com/ishubhamsingh2e/bourbon/jobs/inprocess"
)

func SetupWorker(ctx context.Context) {
	queue := inprocess.New()

	// Configure worker pool
	pool := jobs.NewWorkerPool(queue, jobs.WorkerPoolConfig{
		Concurrency: 4,
		Queues:      []string{"default", "emails"},
	})

	// Register job handler
	pool.Register("send_welcome_email", func(ctx context.Context, j jobs.Job) error {
		log.Printf("Sending welcome email with payload: %s", string(j.Payload()))
		return nil
	})

	// Start processing jobs in background
	pool.Start(ctx)

	// Enqueue job with 30-second delay and 3 retries
	job := jobs.NewJob("send_welcome_email", []byte(`{"email":"user@example.com"}`), nil).
		WithQueue("emails").
		WithMaxRetries(3)

	_ = queue.Enqueue(ctx, job, jobs.WithDelay(30*time.Second))
}
```

---

## 3. Cron Task Scheduling

Bourbon includes a standard 5-part cron syntax scheduler (`schedule/cron`):

```go
package schedule_example

import (
	"context"
	"log"

	"github.com/ishubhamsingh2e/bourbon/schedule/cron"
)

func SetupCron(ctx context.Context) {
	scheduler := cron.NewScheduler()

	// Run midnight every day: "0 0 * * *"
	_ = scheduler.Schedule("0 0 * * *", "daily_cleanup", func(ctx context.Context) error {
		log.Println("Running daily data pruning...")
		return nil
	})

	// Run every 15 minutes: "*/15 * * * *"
	_ = scheduler.Schedule("*/15 * * * *", "sync_rates", func(ctx context.Context) error {
		log.Println("Syncing currency rates...")
		return nil
	})

	scheduler.Start(ctx)
}
```

---

## 4. Real-Time Hub & Server-Sent Events (SSE)

Bourbon provides real-time channel broadcasting with `realtime/inprocess` and native SSE endpoints:

```go
package realtime_example

import (
	"context"
	"time"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/realtime"
	"github.com/ishubhamsingh2e/bourbon/realtime/inprocess"
	"github.com/ishubhamsingh2e/bourbon/router"
)

func SetupRealtime(r router.Router) {
	hub := inprocess.New()

	// Broadcast SSE stream to connected browsers on channel "notifications"
	r.Get("/events", realtime.SSEHandler(hub, "notifications", 15*time.Second))

	// Broadcast message from any controller
	r.Post("/broadcast", func(c *core.Context) error {
		msg := realtime.Message{
			Event: "order_created",
			Data:  `{"order_id": 1024, "status": "paid"}`,
		}
		_ = hub.Broadcast(c.Ctx, "notifications", msg)
		return c.JSON(200, map[string]string{"status": "broadcasted"})
	})
}
```
