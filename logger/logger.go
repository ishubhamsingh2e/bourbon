// Package logger defines structured logging abstractions and ports.
package logger

import "context"

// Logger defines the contract for structured application logging.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	WithContext(ctx context.Context) Logger
}
