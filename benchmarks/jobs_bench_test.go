package benchmarks

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/jobs"
	"github.com/ishubhamsingh2e/bourbon/jobs/inprocess"
)

func BenchmarkJobQueueEnqueueDequeue(b *testing.B) {
	q := inprocess.New()
	ctx := context.Background()
	payload := []byte("test_payload")
	job := jobs.NewJob("bench_job", payload, nil)

	b.ReportAllocs()

	for b.Loop() {
		_ = q.Enqueue(ctx, job)
		_, _ = q.Dequeue(ctx, "default")
	}
}
