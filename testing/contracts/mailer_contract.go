package contracts

import (
	"context"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/mail"
)

// VerifyMailerContract executes a standard compliance test suite against any mail.Mailer implementation.
func VerifyMailerContract(t *testing.T, mailer mail.Mailer) {
	ctx := context.Background()

	t.Run("Send_Message", func(t *testing.T) {
		msg := mail.Message{
			From:     "contract_sender@example.com",
			To:       []string{"contract_receiver@example.com"},
			Subject:  "Contract Test Subject",
			TextBody: "Hello from contract test",
		}

		err := mailer.Send(ctx, msg)
		if err != nil {
			t.Fatalf("Send failed: %v", err)
		}
	})
}
