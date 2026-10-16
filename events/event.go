// Package events provides event publishing, dispatching, and subscription interfaces.
package events

import "time"

// Event defines the contract for domain and application events.
type Event interface {
	Name() string
	OccurredAt() time.Time
	Payload() any
}

// BaseEvent is a convenience standard implementation of the Event interface.
type BaseEvent struct {
	EventName string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// NewEvent constructs a BaseEvent with the current timestamp.
func NewEvent(name string, data any) BaseEvent {
	return BaseEvent{
		EventName: name,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// Name returns the event identifier name.
func (e BaseEvent) Name() string {
	return e.EventName
}

// OccurredAt returns when the event was emitted.
func (e BaseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Payload returns the arbitrary event payload data.
func (e BaseEvent) Payload() any {
	return e.Data
}
