# Async Jobs

Bourbon provides a built-in async job dispatcher interface for handling background task processing.

## Overview

The async job system allows you to offload time-consuming operations (email sending, file processing, API calls) to background workers, keeping your HTTP responses fast and responsive.

## AsyncDispatcher Interface

Bourbon defines an `AsyncDispatcher` interface that you can implement with your preferred job queue backend (Redis, RabbitMQ, database-backed queues, etc.).

```go
type AsyncDispatcher interface {
    Dispatch(handler HandlerFunc, payload interface{}) (string, error)
    GetResult(jobID string) (interface{}, error)
}
```

## Setting Up Async Jobs

### 1. Implement the AsyncDispatcher Interface

Create a custom dispatcher that implements the interface:

```go
package jobs

import (
    "github.com/ishubhamsingh2e/bourbon/bourbon/http"
)

type MyJobQueue struct {
    // Your queue implementation (Redis, RabbitMQ, etc.)
}

func (q *MyJobQueue) Dispatch(handler http.HandlerFunc, payload interface{}) (string, error) {
    // Serialize job to queue
    // Return unique job ID
    jobID := generateUniqueID()
    // ... queue the job
    return jobID, nil
}

func (q *MyJobQueue) GetResult(jobID string) (interface{}, error) {
    // Fetch result from storage
    // Return result or error
    return result, nil
}
```

### 2. Register the Dispatcher

In your application initialization:

```go
func main() {
    cmd.SetCustomInit(func(app *core.Application) error {
        // Create your dispatcher instance
        dispatcher := &jobs.MyJobQueue{
            // ... configuration
        }
        
        // Register with the application
        app.AsyncDispatcher = dispatcher
        
        return nil
    })
    
    cmd.Run("./settings.toml")
}
```

## Using Async Jobs

### Dispatch a Job from a Handler

```go
func sendWelcomeEmail(c *http.Context) error {
    // Parse user data
    var user User
    if err := c.BindJSON(&user); err != nil {
        return err
    }
    
    // Dispatch email job
    jobID, err := c.DispatchAsync(emailHandler, user)
    if err != nil {
        return c.JSON(500, http.H{"error": "Failed to queue email"})
    }
    
    return c.JSON(200, http.H{
        "message": "Welcome email queued",
        "job_id": jobID,
    })
}

func emailHandler(c *http.Context) error {
    // This runs in the background
    user := c.Locals("payload").(User)
    
    // Send email
    sendEmail(user.Email, "Welcome!", "Thanks for joining!")
    
    return nil
}
```

### Quick JSON Response for Queued Jobs

Use `DispatchAsyncJSON` for a standard "job queued" response:

```go
func processVideo(c *http.Context) error {
    var video Video
    c.BindJSON(&video)
    
    // Automatically returns: {"status": "queued", "job_id": "..."}
    return c.DispatchAsyncJSON(videoProcessor, video)
}
```

### Check Job Status

```go
func getJobStatus(c *http.Context) error {
    jobID := c.Param("jobID")
    
    result, err := c.GetAsyncResult(jobID)
    if err != nil {
        return c.JSON(404, http.H{"error": "Job not found"})
    }
    
    return c.JSON(200, http.H{
        "job_id": jobID,
        "result": result,
    })
}
```

## Best Practices

1. **Keep Jobs Idempotent**: Design jobs to be safely retried in case of failures.

2. **Use Timeouts**: Implement reasonable timeouts in your dispatcher to prevent jobs from hanging.

3. **Error Handling**: Always handle errors in background jobs and log them appropriately.

4. **Payload Serialization**: Ensure your payloads can be properly serialized/deserialized by your queue backend.

5. **Resource Cleanup**: Clean up resources (file handles, connections) in your job handlers.

## Example: Redis-Backed Dispatcher

Here's a simple example using Redis as a job queue:

```go
package jobs

import (
    "encoding/json"
    "github.com/go-redis/redis/v8"
    "github.com/ishubhamsingh2e/bourbon/bourbon/http"
    "github.com/google/uuid"
)

type RedisDispatcher struct {
    client *redis.Client
}

func NewRedisDispatcher(addr string) *RedisDispatcher {
    return &RedisDispatcher{
        client: redis.NewClient(&redis.Options{
            Addr: addr,
        }),
    }
}

func (r *RedisDispatcher) Dispatch(handler http.HandlerFunc, payload interface{}) (string, error) {
    jobID := uuid.New().String()
    
    job := Job{
        ID:      jobID,
        Payload: payload,
    }
    
    data, err := json.Marshal(job)
    if err != nil {
        return "", err
    }
    
    // Push to Redis queue
    ctx := context.Background()
    err = r.client.LPush(ctx, "jobs:queue", data).Err()
    if err != nil {
        return "", err
    }
    
    return jobID, nil
}

func (r *RedisDispatcher) GetResult(jobID string) (interface{}, error) {
    ctx := context.Background()
    result, err := r.client.Get(ctx, "jobs:result:"+jobID).Result()
    if err != nil {
        return nil, err
    }
    
    var data interface{}
    json.Unmarshal([]byte(result), &data)
    return data, nil
}
```

## Future Enhancements

The async job system is designed to be extensible. Planned features include:

- Built-in database-backed queue implementation
- Job scheduling and cron support
- Worker pool management utilities
- Job retry policies and exponential backoff
- Dead letter queue handling
- Job priority levels

## See Also

- [Requests & Responses](requests_responses.md) - Context methods and data binding
- [Middleware](middleware.md) - Processing requests before handlers
