package app_test

import (
	"context"
	"errors"
	"net/smtp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

type fakeDialer struct {
	addr string
	from string
	to   []string
	msg  []byte
	err  error
}

func (d *fakeDialer) SendMail(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
	d.addr, d.from, d.to, d.msg = addr, from, to, msg
	return d.err
}

func TestSMTPSender_BuildsAndDispatches(t *testing.T) {
	dialer := &fakeDialer{}
	sender := app.NewSMTPSenderWithDialer(app.SMTPConfig{Addr: "smtp:25", From: "no-reply@forge"}, dialer)

	assert.Equal(t, domain.ChannelEmail, sender.Channel())

	n := internaltest.NewNotification(
		internaltest.WithNRecipient("user@example.com"),
		internaltest.WithNSubject("Verify"),
		internaltest.WithNBody("click here"),
	)
	require.NoError(t, sender.Send(context.Background(), n))

	assert.Equal(t, "smtp:25", dialer.addr)
	assert.Equal(t, []string{"user@example.com"}, dialer.to)
	body := string(dialer.msg)
	assert.True(t, strings.Contains(body, "Subject: Verify"))
	assert.True(t, strings.Contains(body, "click here"))
}

func TestSMTPSender_PropagatesError(t *testing.T) {
	dialer := &fakeDialer{err: errors.New("connection refused")}
	sender := app.NewSMTPSenderWithDialer(app.SMTPConfig{Addr: "smtp:25"}, dialer)
	require.Error(t, sender.Send(context.Background(), internaltest.NewNotification()))
}

// TestSMTPSender_ConfirmsDelivery pins that an SMTP send, which returns nil only
// after the receiving server accepts the message, reports a confirmed
// StatusDelivered via the ConfirmingSender capability.
func TestSMTPSender_ConfirmsDelivery(t *testing.T) {
	sender := app.NewSMTPSenderWithDialer(app.SMTPConfig{Addr: "smtp:25"}, &fakeDialer{})
	confirming, ok := any(sender).(app.ConfirmingSender)
	require.True(t, ok, "SMTPSender must implement ConfirmingSender")
	assert.Equal(t, domain.StatusDelivered, confirming.SuccessStatus())
}
