// Package slogadapter provides a standard library log/slog implementation of logger.Logger.
package slogadapter

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/ishubhamsingh2e/bourbon/logger"
)

// Adapter wraps standard log/slog.Logger to satisfy logger.Logger.
type Adapter struct {
	l   *slog.Logger
	ctx context.Context
}

// New constructs an Adapter writing to w.
func New(w io.Writer, jsonFormat bool, level slog.Level) *Adapter {
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if jsonFormat {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return &Adapter{
		l:   slog.New(handler),
		ctx: context.Background(),
	}
}

// NewDefault constructs an Adapter writing text logs to os.Stdout at Info level.
func NewDefault() *Adapter {
	return New(os.Stdout, false, slog.LevelInfo)
}

// Debug logs at debug level.
func (a *Adapter) Debug(msg string, args ...any) {
	a.l.DebugContext(a.ctx, msg, args...)
}

// Info logs at info level.
func (a *Adapter) Info(msg string, args ...any) {
	a.l.InfoContext(a.ctx, msg, args...)
}

// Warn logs at warn level.
func (a *Adapter) Warn(msg string, args ...any) {
	a.l.WarnContext(a.ctx, msg, args...)
}

// Error logs at error level.
func (a *Adapter) Error(msg string, args ...any) {
	a.l.ErrorContext(a.ctx, msg, args...)
}

// With returns a new Logger with bound structured attributes.
func (a *Adapter) With(args ...any) logger.Logger {
	return &Adapter{
		l:   a.l.With(args...),
		ctx: a.ctx,
	}
}

// WithContext returns a new Logger scoped with the given context.
func (a *Adapter) WithContext(ctx context.Context) logger.Logger {
	return &Adapter{
		l:   a.l,
		ctx: ctx,
	}
}
