// Package mail provides email abstractions, message formatting, and delivery drivers.
package mail

import "context"

// Message represents an email to be sent.
type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	TextBody    string
	HTMLBody    string
	Attachments map[string][]byte
	Headers     map[string]string
}

// Mailer defines the contract for sending email messages.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}
