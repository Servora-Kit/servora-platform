package mail

import (
	"context"
	"crypto/tls"
	"fmt"

	conf "github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	gomail "github.com/wneessen/go-mail"
)

type smtpSender struct {
	host string
	opts []gomail.Option
}

func newSMTPSender(c *conf.Smtp) *smtpSender {
	port := int(c.Port)
	if port == 0 {
		port = 587
	}

	opts := []gomail.Option{
		gomail.WithPort(port),
	}

	if c.Username != "" || c.Password != "" {
		opts = append(opts,
			gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
			gomail.WithUsername(c.Username),
			gomail.WithPassword(c.Password),
		)
	}

	if c.Tls {
		opts = append(opts, gomail.WithTLSPortPolicy(gomail.TLSMandatory))
	} else {
		opts = append(opts, gomail.WithTLSPortPolicy(gomail.TLSOpportunistic))
	}

	if c.SkipVerifySsl {
		opts = append(opts, gomail.WithTLSConfig(&tls.Config{
			ServerName:         c.Host,
			InsecureSkipVerify: true, //nolint:gosec // configurable for dev
		}))
	}

	if c.SendTimeout != nil {
		d := c.SendTimeout.AsDuration()
		if d > 0 {
			opts = append(opts, gomail.WithTimeout(d))
		}
	}

	return &smtpSender{
		host: c.Host,
		opts: opts,
	}
}

func (s *smtpSender) Send(ctx context.Context, email Email) error {
	msg := gomail.NewMsg()

	fromStr := fmt.Sprintf(`"%s" <%s>`, email.From.Name, email.From.Address)
	if err := msg.From(fromStr); err != nil {
		return fmt.Errorf("mail: set From header: %w", err)
	}

	if len(email.To) > 0 {
		if err := msg.To(email.To...); err != nil {
			return fmt.Errorf("mail: set To header: %w", err)
		}
	}

	if len(email.Cc) > 0 {
		if err := msg.Cc(email.Cc...); err != nil {
			return fmt.Errorf("mail: set Cc header: %w", err)
		}
	}

	if len(email.Bcc) > 0 {
		if err := msg.Bcc(email.Bcc...); err != nil {
			return fmt.Errorf("mail: set Bcc header: %w", err)
		}
	}

	msg.Subject(email.Subject)

	if len(email.HTML) > 0 {
		msg.SetBodyString(gomail.TypeTextHTML, string(email.HTML))
		if len(email.Text) > 0 {
			msg.AddAlternativeString(gomail.TypeTextPlain, string(email.Text))
		}
	} else if len(email.Text) > 0 {
		msg.SetBodyString(gomail.TypeTextPlain, string(email.Text))
	}

	client, err := gomail.NewClient(s.host, s.opts...)
	if err != nil {
		return fmt.Errorf("mail: create SMTP client: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mail: send: %w", err)
	}
	return nil
}

// ensure smtpSender satisfies Sender at compile time.
var _ Sender = (*smtpSender)(nil)
