// Package inprocess provides an in-memory channel-based implementation of events.Bus.
package inprocess

import (
	"context"
	"errors"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/events"
)

type asyncTask struct {
	ctx      context.Context
	event    events.Event
	listener events.ListenerFunc
}

// Bus is a thread-safe, in-process event dispatcher.
type Bus struct {
	mu             sync.RWMutex
	listeners      map[string][]events.ListenerFunc
	asyncListeners map[string][]events.ListenerFunc
	queue          chan asyncTask
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// Config configures in-process event bus settings.
type Config struct {
	BufferSize int
}

// Open satisfies the events.Driver interface.
func (c Config) Open() (events.Bus, error) {
	buf := c.BufferSize
	if buf <= 0 {
		buf = 1024
	}
	return New(buf), nil
}

// New constructs an in-process Bus with worker goroutines for background event publishing.
func New(bufferSize ...int) *Bus {
	buf := 1024
	if len(bufferSize) > 0 && bufferSize[0] > 0 {
		buf = bufferSize[0]
	}

	ctx, cancel := context.WithCancel(context.Background())
	b := &Bus{
		listeners:      make(map[string][]events.ListenerFunc),
		asyncListeners: make(map[string][]events.ListenerFunc),
		queue:          make(chan asyncTask, buf),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Spawn worker pool
	for i := 0; i < 4; i++ {
		b.wg.Add(1)
		go b.worker()
	}

	return b
}

func (b *Bus) worker() {
	defer b.wg.Done()
	for {
		select {
		case <-b.ctx.Done():
			// Drain remaining tasks before exit
			for {
				select {
				case task := <-b.queue:
					_ = task.listener(task.ctx, task.event)
				default:
					return
				}
			}
		case task, ok := <-b.queue:
			if !ok {
				return
			}
			_ = task.listener(task.ctx, task.event)
		}
	}
}

// Close gracefully drains pending async event listeners and shuts down the bus.
func (b *Bus) Close() error {
	b.cancel()
	b.wg.Wait()
	return nil
}

// Subscribe registers a synchronous listener for the given event name or "*" wildcard.
func (b *Bus) Subscribe(eventName string, listener events.ListenerFunc) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.listeners[eventName] = append(b.listeners[eventName], listener)

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.listeners[eventName]
		for i, l := range list {
			// Compare function pointers
			if &l == &listener {
				b.listeners[eventName] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}
}

// SubscribeAsync registers an asynchronous listener executed in the worker pool.
func (b *Bus) SubscribeAsync(eventName string, listener events.ListenerFunc) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.asyncListeners[eventName] = append(b.asyncListeners[eventName], listener)

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.asyncListeners[eventName]
		for i, l := range list {
			if &l == &listener {
				b.asyncListeners[eventName] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}
}

// Dispatch executes all registered synchronous listeners in-line and returns any accumulated errors.
func (b *Bus) Dispatch(ctx context.Context, e events.Event) error {
	b.mu.RLock()
	var targets []events.ListenerFunc
	if list, ok := b.listeners[e.Name()]; ok {
		targets = append(targets, list...)
	}
	if list, ok := b.listeners["*"]; ok {
		targets = append(targets, list...)
	}
	b.mu.RUnlock()

	var errs []error
	for _, fn := range targets {
		if err := fn(ctx, e); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Publish enqueues async listeners to the background worker pool and executes synchronous listeners in-line.
func (b *Bus) Publish(ctx context.Context, e events.Event) error {
	b.mu.RLock()
	var asyncTargets []events.ListenerFunc
	if list, ok := b.asyncListeners[e.Name()]; ok {
		asyncTargets = append(asyncTargets, list...)
	}
	if list, ok := b.asyncListeners["*"]; ok {
		asyncTargets = append(asyncTargets, list...)
	}
	b.mu.RUnlock()

	// Enqueue async listeners
	for _, fn := range asyncTargets {
		select {
		case b.queue <- asyncTask{ctx: ctx, event: e, listener: fn}:
		default:
			// If buffer is full, execute in a standalone goroutine to guarantee delivery
			go func(l events.ListenerFunc) {
				_ = l(ctx, e)
			}(fn)
		}
	}

	return b.Dispatch(ctx, e)
}

// Provider implements core.Provider for in-process event bus integration.
type Provider struct {
	bus *Bus
}

// NewProvider returns a new in-process event bus Provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Register registers events.Bus into the DI container as a singleton.
func (p *Provider) Register(c core.Container) error {
	p.bus = New()
	c.Singleton(func() events.Bus {
		return p.bus
	})
	return nil
}

// Boot verifies event bus readiness.
func (p *Provider) Boot(ctx context.Context, app core.Application) error {
	var b events.Bus
	return app.Container().Resolve(&b)
}

// Shutdown gracefully stops the event bus.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.bus != nil {
		return p.bus.Close()
	}
	return nil
}

var _ core.Provider = (*Provider)(nil)
