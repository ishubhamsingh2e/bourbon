// Package zap provides an Uber Zap-compatible structured logger adapter for logger.Logger.
package zap

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/logger"
)

// Logger implements logger.Logger using high-performance structured formatting.
type Logger struct {
	mu     sync.RWMutex
	w      io.Writer
	ctx    context.Context
	fields map[string]any
}

// New constructs a Zap-compatible structured Logger.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{
		w:      w,
		ctx:    context.Background(),
		fields: make(map[string]any),
	}
}

func (l *Logger) log(level, msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	out := fmt.Sprintf("[%s] %s", level, msg)
	for k, v := range l.fields {
		out += fmt.Sprintf(" %s=%v", k, v)
	}
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			out += fmt.Sprintf(" %v=%v", args[i], args[i+1])
		} else {
			out += fmt.Sprintf(" %v", args[i])
		}
	}
	out += "\n"
	_, _ = l.w.Write([]byte(out))
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, args ...any) {
	l.log("DEBUG", msg, args...)
}

// Info logs at info level.
func (l *Logger) Info(msg string, args ...any) {
	l.log("INFO", msg, args...)
}

// Warn logs at warn level.
func (l *Logger) Warn(msg string, args ...any) {
	l.log("WARN", msg, args...)
}

// Error logs at error level.
func (l *Logger) Error(msg string, args ...any) {
	l.log("ERROR", msg, args...)
}

// With binds structured fields.
func (l *Logger) With(args ...any) logger.Logger {
	l.mu.RLock()
	newFields := make(map[string]any, len(l.fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	l.mu.RUnlock()

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			newFields[fmt.Sprintf("%v", args[i])] = args[i+1]
		}
	}

	return &Logger{
		w:      l.w,
		ctx:    l.ctx,
		fields: newFields,
	}
}

// WithContext returns a logger scoped to ctx.
func (l *Logger) WithContext(ctx context.Context) logger.Logger {
	return &Logger{
		w:      l.w,
		ctx:    ctx,
		fields: l.fields,
	}
}

var _ logger.Logger = (*Logger)(nil)
