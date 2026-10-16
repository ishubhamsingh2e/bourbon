package events

import "context"

// ListenerFunc defines the signature of an event handler.
type ListenerFunc func(ctx context.Context, e Event) error

// Bus defines the contract for dispatching and subscribing to events.
type Bus interface {
	// Publish asynchronously dispatches an event to registered listeners without blocking the caller.
	Publish(ctx context.Context, e Event) error

	// Dispatch synchronously invokes all registered listeners in order and returns any accumulated errors.
	Dispatch(ctx context.Context, e Event) error

	// Subscribe registers a synchronous listener for the given event name (or "*" for all events).
	Subscribe(eventName string, listener ListenerFunc) (unsubscribe func())

	// SubscribeAsync registers an asynchronous listener for the given event name.
	SubscribeAsync(eventName string, listener ListenerFunc) (unsubscribe func())
}

// Driver represents an event bus provider capable of opening an event Bus instance.
type Driver interface {
	Open() (Bus, error)
}
