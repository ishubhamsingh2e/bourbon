package devmail

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/mail"
)

func TestDevMailerOperations(t *testing.T) {
	mailer := New()
	ctx := context.Background()

	msg := mail.Message{
		From:    "sender@example.com",
		To:      []string{"receiver@example.com"},
		Subject: "Test Email",
	}

	err := mailer.Send(ctx, msg)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	sent := mailer.SentMessages()
	if len(sent) != 1 {
		t.Fatalf("expected 1 sent message, got: %d", len(sent))
	}
	if sent[0].Subject != "Test Email" {
		t.Errorf("subject mismatch: %s", sent[0].Subject)
	}

	mailer.Reset()
	if len(mailer.SentMessages()) != 0 {
		t.Errorf("expected 0 messages after reset")
	}
}

func TestDevMailProviderLifecycle(t *testing.T) {
	provider := NewProvider()
	application := app.New()

	if err := provider.Register(application.Container()); err != nil {
		t.Fatalf("failed registering devmail provider: %v", err)
	}

	var m mail.Mailer
	if err := application.Container().Resolve(&m); err != nil {
		t.Fatalf("failed resolving mail.Mailer: %v", err)
	}

	ctx := context.Background()
	if err := provider.Boot(ctx, application); err != nil {
		t.Fatalf("failed booting devmail provider: %v", err)
	}

	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("failed shutting down devmail provider: %v", err)
	}
}
