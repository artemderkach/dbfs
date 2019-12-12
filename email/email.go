// Package email provides functionality for sending email via email API service
package email

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v3"
)

// Email stucture provides access for sending emails
type Email struct {
	mailgun       *mailgun.MailgunImpl
	mailgunDomain string
}

// New creates Email instance. Needed for encapsulating API key, domain
func New(mailgunAPIKey, mailgunDomain string) *Email {
	mg := mailgun.NewMailgun(mailgunDomain, mailgunAPIKey)
	return &Email{mailgun: mg, mailgunDomain: mailgunDomain}
}

// Send sends email to specific address
func (e *Email) Send(targetEmail, msgBody string) (string, error) {
	m := e.mailgun.NewMessage(
		"dbfs@"+e.mailgunDomain,
		"token",
		msgBody,
		targetEmail,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	resp, _, err := e.mailgun.Send(ctx, m)
	return resp, err
}
