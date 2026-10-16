// Package devmail provides an in-memory and logger-based implementation of mail.Mailer for development and tests.
package devmail

import (
	"context"
	"log/slog"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/mail"
)

// Mailer records sent emails in memory and logs them.
type Mailer struct {
	mu   sync.RWMutex
	sent []mail.Message
}

// New constructs a DevMail mailer.
func New() *Mailer {
	return &Mailer{
		sent: make([]mail.Message, 0),
	}
}

// Send records the email and logs its delivery.
func (m *Mailer) Send(ctx context.Context, msg mail.Message) error {
	m.mu.Lock()
	m.sent = append(m.sent, msg)
	m.mu.Unlock()

	slog.InfoContext(ctx, "devmail: sent email",
		"from", msg.From,
		"to", msg.To,
		"subject", msg.Subject,
	)
	return nil
}

// SentMessages returns a copy of all emails sent through this mailer.
func (m *Mailer) SentMessages() []mail.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]mail.Message, len(m.sent))
	copy(result, m.sent)
	return result
}

// Reset clears the recorded sent emails.
func (m *Mailer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = make([]mail.Message, 0)
}

// Provider implements core.Provider for development mailer integration.
type Provider struct {
	mailer *Mailer
}

// NewProvider returns a new DevMail Provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Register registers mail.Mailer into the DI container as a singleton.
func (p *Provider) Register(c core.Container) error {
	p.mailer = New()
	c.Singleton(func() mail.Mailer {
		return p.mailer
	})
	return nil
}

// Boot verifies mailer readiness.
func (p *Provider) Boot(ctx context.Context, app core.Application) error {
	var m mail.Mailer
	return app.Container().Resolve(&m)
}

// Shutdown releases mailer resources.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.mailer != nil {
		p.mailer.Reset()
	}
	return nil
}

var _ core.Provider = (*Provider)(nil)
