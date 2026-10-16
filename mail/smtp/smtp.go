// Package smtp provides a standard library net/smtp implementation of mail.Mailer.
package smtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/ishubhamsingh2e/bourbon/mail"
)

// Config configures an SMTP connection.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Mailer implements mail.Mailer using net/smtp.
type Mailer struct {
	cfg Config
}

// New constructs an SMTP mailer.
func New(cfg Config) *Mailer {
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	return &Mailer{cfg: cfg}
}

// BuildRawMessage constructs an RFC 822 / MIME multipart formatted byte slice.
func (m *Mailer) BuildRawMessage(msg mail.Message) ([]byte, error) {
	from := msg.From
	if from == "" {
		from = m.cfg.From
	}
	if from == "" {
		return nil, errors.New("smtp: missing From address")
	}
	if len(msg.To) == 0 {
		return nil, errors.New("smtp: missing To recipients")
	}

	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	if len(msg.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.Cc, ", ")))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))

	for k, v := range msg.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// Simple single-part message
	if len(msg.Attachments) == 0 && msg.HTMLBody == "" {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		buf.WriteString(msg.TextBody)
		return buf.Bytes(), nil
	}

	// Multipart message
	writer := multipart.NewWriter(&buf)
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", writer.Boundary()))

	// Text body part
	if msg.TextBody != "" {
		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Type", "text/plain; charset=utf-8")
		part, err := writer.CreatePart(partHeader)
		if err != nil {
			return nil, err
		}
		_, _ = part.Write([]byte(msg.TextBody))
	}

	// HTML body part
	if msg.HTMLBody != "" {
		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Type", "text/html; charset=utf-8")
		part, err := writer.CreatePart(partHeader)
		if err != nil {
			return nil, err
		}
		_, _ = part.Write([]byte(msg.HTMLBody))
	}

	// Attachments
	for filename, data := range msg.Attachments {
		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Type", "application/octet-stream")
		partHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		partHeader.Set("Content-Transfer-Encoding", "base64")

		part, err := writer.CreatePart(partHeader)
		if err != nil {
			return nil, err
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		_, _ = part.Write([]byte(encoded))
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Send transmits the email over SMTP.
func (m *Mailer) Send(ctx context.Context, msg mail.Message) error {
	raw, err := m.BuildRawMessage(msg)
	if err != nil {
		return err
	}

	from := msg.From
	if from == "" {
		from = m.cfg.From
	}

	var recipients []string
	recipients = append(recipients, msg.To...)
	recipients = append(recipients, msg.Cc...)
	recipients = append(recipients, msg.Bcc...)

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	var auth smtp.Auth
	if m.cfg.Username != "" && m.cfg.Password != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	return smtp.SendMail(addr, auth, from, recipients, raw)
}
