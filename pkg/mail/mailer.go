package mail

import (
	"context"
	"fmt"

	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
)

type From struct {
	Address string
	Name    string
}

type Email struct {
	From    From
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Text    []byte
	HTML    []byte
}

type Sender interface {
	Send(ctx context.Context, email Email) error
}

type noopSender struct{}

func (noopSender) Send(context.Context, Email) error {
	return fmt.Errorf("mail: sender not configured (no mail config provided)")
}

func NewSender(c *conf.Mail) Sender {
	if c == nil || c.Smtp == nil {
		return noopSender{}
	}
	return newSMTPSender(c.Smtp)
}

func DefaultFrom(c *conf.Mail) From {
	if c == nil || c.From == nil {
		return From{}
	}
	return From{
		Address: c.From.Address,
		Name:    c.From.Name,
	}
}
