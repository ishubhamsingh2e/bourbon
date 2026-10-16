package benchmarks

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/events"
	"github.com/ishubhamsingh2e/bourbon/events/inprocess"
)

func BenchmarkEventBusSyncDispatch(b *testing.B) {
	bus := inprocess.New(100)
	defer bus.Close()

	bus.Subscribe("user.event", func(ctx context.Context, e events.Event) error {
		return nil
	})

	ctx := context.Background()
	e := events.NewEvent("user.event", []byte("payload"))

	b.ReportAllocs()

	for b.Loop() {
		_ = bus.Dispatch(ctx, e)
	}
}

func BenchmarkEventBusAsyncPublish(b *testing.B) {
	bus := inprocess.New(50000)
	defer bus.Close()

	bus.Subscribe("user.async_event", func(ctx context.Context, e events.Event) error {
		return nil
	})

	ctx := context.Background()
	e := events.NewEvent("user.async_event", []byte("payload"))

	b.ReportAllocs()

	for b.Loop() {
		_ = bus.Publish(ctx, e)
	}
}
