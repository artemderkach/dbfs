// Package email provides functionality for sending email via email API service
package email

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v3"
)

// EmailService interface needed for testing purposes
type EmailService interface {
	Send(targetEmail, msgBody string) (string, error)
}

// Email stucture provides access for sending emails
type Email struct {
	mailgun           *mailgun.MailgunImpl
	mailgunRootDomain string
	mailgunSubdomain  string
}

// New creates Email instance. Needed for encapsulating API key, domain
func New(mailgunAPIKey, mailgunRootDomain, mailgunSubdomain string) *Email {
	mg := mailgun.NewMailgun(mailgunSubdomain, mailgunAPIKey)
	mg.SetAPIBase(mailgun.APIBaseEU)
	return &Email{
		mailgun:           mg,
		mailgunRootDomain: mailgunRootDomain,
		mailgunSubdomain:  mailgunSubdomain,
	}
}

// Send sends email to specific address
func (e *Email) Send(targetEmail, msgBody string) (string, error) {
	m := e.mailgun.NewMessage(
		"dbfs@"+e.mailgunRootDomain,
		"token",
		msgBody,
		targetEmail,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	resp, _, err := e.mailgun.Send(ctx, m)
	return resp, err
}
