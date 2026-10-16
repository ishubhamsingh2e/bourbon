# File Storage & Mailer Subsystems in Bourbon

Bourbon abstracts file persistence and email delivery behind clean port interfaces (`storage.Storage` and `mail.Mailer`), with native standard library drivers and swappable cloud adapters.

---

## 1. File Storage (`storage.Storage`)

Bourbon provides the `storage.Storage` port with methods for saving, retrieving, inspecting, and deleting binary objects:

```go
package uploads

import (
	"context"
	"io"

	"github.com/ishubhamsingh2e/bourbon/storage"
	"github.com/ishubhamsingh2e/bourbon/storage/localfs"
)

func SaveAvatar(ctx context.Context, fileReader io.Reader) (string, error) {
	// Initialize local filesystem storage (rootDir, publicBaseURL)
	store, err := localfs.New("storage/app/public", "http://localhost:8000/storage")
	if err != nil {
		return "", err
	}

	path := "avatars/user_1024.png"

	// 1. Write file
	if err := store.Put(ctx, path, fileReader); err != nil {
		return "", err
	}

	// 2. Check existence & size
	exists, _ := store.Exists(ctx, path)
	size, _ := store.Size(ctx, path)

	// 3. Obtain public URL
	publicURL, err := store.URL(ctx, path)
	return publicURL, err
}
```

### Available Drivers
- **Local Filesystem**: `github.com/ishubhamsingh2e/bourbon/storage/localfs` (standard Go os/io).
- **S3 / Cloudflare R2**: `github.com/ishubhamsingh2e/bourbon/contrib/s3` (S3 API compatibility).

---

## 2. Mailer Subsystem (`mail.Mailer`)

Bourbon provides standard email delivery with full support for HTML templates, plain text fallback, CC/BCC, and attachments:

```go
package notifications

import (
	"context"

	"github.com/ishubhamsingh2e/bourbon/mail"
	"github.com/ishubhamsingh2e/bourbon/mail/smtp"
)

func SendWelcomeEmail(ctx context.Context, toEmail, name string, pdfInvoice []byte) error {
	mailer := smtp.New(smtp.Config{
		Host:     "smtp.mailgun.org",
		Port:     587,
		Username: "postmaster@example.com",
		Password: "secret-password",
		From:     "welcome@example.com",
	})

	msg := mail.Message{
		From:     "welcome@example.com",
		To:       []string{toEmail},
		Subject:  "Welcome to Bourbon!",
		TextBody: "Hello " + name + ", thank you for joining.",
		HTMLBody: "<h1>Welcome, " + name + "!</h1><p>Thank you for joining.</p>",
		Attachments: map[string][]byte{
			"receipt.pdf": pdfInvoice,
		},
	}

	return mailer.Send(ctx, msg)
}
```

### Testing with `mail/devmail`

In unit tests and local development, swap `smtp` for `mail/devmail`, which captures sent emails in memory and logs them via `slog` without making network connections:

```go
package notifications_test

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/mail/devmail"
)

func TestEmailDelivery(t *testing.T) {
	mailer := devmail.New()

	// Perform action that triggers email...

	messages := mailer.SentMessages()
	if len(messages) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(messages))
	}
	if messages[0].To[0] != "user@example.com" {
		t.Errorf("unexpected recipient: %s", messages[0].To[0])
	}
}
```
