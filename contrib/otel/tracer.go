// Package otel provides OpenTelemetry distributed tracing adapters and HTTP middleware.
package otel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type traceContextKey struct{}

// Span represents an in-flight OpenTelemetry tracing span.
type Span struct {
	TraceID    string
	SpanID     string
	ParentID   string
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	Attributes map[string]any
}

// Tracer manages trace contexts and spans.
type Tracer struct {
	serviceName string
	mu          sync.RWMutex
	spans       []*Span
}

// NewTracer constructs a Tracer for the given service.
func NewTracer(serviceName string) *Tracer {
	return &Tracer{
		serviceName: serviceName,
		spans:       make([]*Span, 0),
	}
}

// Start creates a new span and injects it into the returned context.
func (t *Tracer) Start(ctx context.Context, spanName string) (context.Context, *Span) {
	parentSpan, _ := ctx.Value(traceContextKey{}).(*Span)

	traceID := generateHex(16)
	parentID := ""
	if parentSpan != nil {
		traceID = parentSpan.TraceID
		parentID = parentSpan.SpanID
	}

	spanID := generateHex(8)

	span := &Span{
		TraceID:    traceID,
		SpanID:     spanID,
		ParentID:   parentID,
		Name:       spanName,
		StartTime:  time.Now(),
		Attributes: make(map[string]any),
	}

	t.mu.Lock()
	t.spans = append(t.spans, span)
	t.mu.Unlock()

	return context.WithValue(ctx, traceContextKey{}, span), span
}

// End completes the span.
func (s *Span) End() {
	s.EndTime = time.Now()
}

// SetAttribute assigns a key-value attribute to the span.
func (s *Span) SetAttribute(key string, val any) {
	s.Attributes[key] = val
}

// TraceParentHeader returns the standard W3C Trace Context traceparent header value.
func (s *Span) TraceParentHeader() string {
	return fmt.Sprintf("00-%s-%s-01", s.TraceID, s.SpanID)
}

func generateHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
