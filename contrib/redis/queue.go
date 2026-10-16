package redis

import (
	"context"
	"encoding/json"

	"github.com/ishubhamsingh2e/bourbon/jobs"
)

type serializedJob struct {
	Name    string `json:"name"`
	Payload []byte `json:"payload"`
}

// Queue implements jobs.Queue on top of Redis lists.
type Queue struct {
	client Client
	prefix string
}

// NewQueue constructs a Redis job queue adapter.
func NewQueue(client Client, prefix ...string) *Queue {
	p := "bourbon:jobs:"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &Queue{client: client, prefix: p}
}

func (q *Queue) queueKey(name string) string {
	return q.prefix + name
}

// Enqueue serializes and pushes a job to Redis list.
func (q *Queue) Enqueue(ctx context.Context, j jobs.Job, opts ...jobs.JobOption) error {
	options := jobs.DefaultJobOptions()
	for _, opt := range opts {
		opt(&options)
	}

	payload, err := json.Marshal(serializedJob{
		Name:    j.Name(),
		Payload: j.Payload(),
	})
	if err != nil {
		return err
	}

	return q.client.RPush(ctx, q.queueKey(options.Queue), payload)
}

// Dequeue pops the next job from Redis list.
func (q *Queue) Dequeue(ctx context.Context, queues ...string) (jobs.Job, error) {
	if len(queues) == 0 {
		queues = []string{"default"}
	}

	for _, queueName := range queues {
		bytes, ok, err := q.client.LPop(ctx, q.queueKey(queueName))
		if err != nil {
			return nil, err
		}
		if ok {
			var sj serializedJob
			if err := json.Unmarshal(bytes, &sj); err != nil {
				return nil, err
			}
			return jobs.NewJob(sj.Name, sj.Payload, nil), nil
		}
	}

	return nil, jobs.ErrQueueEmpty
}

// Size returns the length of the specified queue.
func (q *Queue) Size(ctx context.Context, queues ...string) (int64, error) {
	if len(queues) == 0 {
		queues = []string{"default"}
	}

	var total int64
	for _, queueName := range queues {
		len, err := q.client.LLen(ctx, q.queueKey(queueName))
		if err != nil {
			return 0, err
		}
		total += len
	}

	return total, nil
}
