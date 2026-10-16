package zap

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestZapAdapterLogging(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf).With("service", "payment")
	ctx := context.Background()

	l.WithContext(ctx).Info("charge initiated", "amount", 99.99)

	out := buf.String()
	if !strings.Contains(out, "[INFO] charge initiated") || !strings.Contains(out, "service=payment") || !strings.Contains(out, "amount=99.99") {
		t.Errorf("log output mismatch: %s", out)
	}
}
