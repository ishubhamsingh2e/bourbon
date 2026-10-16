package smtp

import (
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/mail"
)

func TestBuildRawMessage(t *testing.T) {
	mailer := New(Config{
		Host: "smtp.example.com",
		Port: 587,
		From: "no-reply@example.com",
	})

	msg := mail.Message{
		To:          []string{"alice@example.com"},
		Subject:     "Welcome to Bourbon",
		TextBody:    "Hello Alice",
		HTMLBody:    "<h1>Hello Alice</h1>",
		Attachments: map[string][]byte{"invoice.pdf": []byte("fake_pdf_data")},
	}

	raw, err := mailer.BuildRawMessage(msg)
	if err != nil {
		t.Fatalf("BuildRawMessage failed: %v", err)
	}

	rawStr := string(raw)
	if !strings.Contains(rawStr, "From: no-reply@example.com") {
		t.Errorf("missing From header in raw email: %s", rawStr)
	}
	if !strings.Contains(rawStr, "To: alice@example.com") {
		t.Errorf("missing To header in raw email: %s", rawStr)
	}
	if !strings.Contains(rawStr, "Subject: Welcome to Bourbon") {
		t.Errorf("missing Subject header in raw email: %s", rawStr)
	}
	if !strings.Contains(rawStr, "filename=\"invoice.pdf\"") {
		t.Errorf("missing attachment in raw email: %s", rawStr)
	}
}
