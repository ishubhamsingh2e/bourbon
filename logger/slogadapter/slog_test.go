package slogadapter

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestSlogAdapterTextAndJSON(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, true, slog.LevelDebug)

	l.Info("server starting", "port", 8080)
	out := buf.String()

	if !strings.Contains(out, `"msg":"server starting"`) || !strings.Contains(out, `"port":8080`) {
		t.Errorf("JSON log mismatch: %s", out)
	}

	// With attributes
	var buf2 bytes.Buffer
	l2 := New(&buf2, false, slog.LevelInfo).With("service", "auth")
	l2.WithContext(context.Background()).Info("token issued")

	out2 := buf2.String()
	if !strings.Contains(out2, "token issued") || !strings.Contains(out2, "service=auth") {
		t.Errorf("Text log mismatch: %s", out2)
	}
}
