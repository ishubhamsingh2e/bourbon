package realtime

// Client defines the contract for an active real-time subscriber connection.
type Client interface {
	// ID returns a unique identifier for this client connection.
	ID() string

	// Send transmits a message to the client.
	Send(msg Message) error

	// Close terminates the client connection.
	Close() error
}
