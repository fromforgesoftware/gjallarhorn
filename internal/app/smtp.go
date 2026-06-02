package app

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// MailDialer is the seam over net/smtp so tests need no real SMTP server.
// SES/Mailgun adapters implement ChannelSender directly rather than this.
type MailDialer interface {
	SendMail(addr string, auth smtp.Auth, from string, to []string, msg []byte) error
}

type netSMTPDialer struct{}

func (netSMTPDialer) SendMail(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, auth, from, to, msg)
}

// SMTPConfig configures the EMAIL channel's SMTP transport.
type SMTPConfig struct {
	Addr     string `envconfig:"SMTP_ADDR"`
	From     string `envconfig:"SMTP_FROM"`
	Username string `envconfig:"SMTP_USERNAME"`
	Password string `envconfig:"SMTP_PASSWORD"`
	Host     string `envconfig:"SMTP_HOST"`
}

type SMTPSender struct {
	cfg    SMTPConfig
	dialer MailDialer
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg, dialer: netSMTPDialer{}}
}

func NewSMTPSenderWithDialer(cfg SMTPConfig, dialer MailDialer) *SMTPSender {
	return &SMTPSender{cfg: cfg, dialer: dialer}
}

func (s *SMTPSender) Channel() string { return domain.ChannelEmail }

func (s *SMTPSender) Send(_ context.Context, n domain.Notification) error {
	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}
	msg := buildMessage(s.cfg.From, n.Recipient(), n.Subject(), n.Body())
	return s.dialer.SendMail(s.cfg.Addr, auth, s.cfg.From, []string{n.Recipient()}, msg)
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}
