package app

import (
	"context"

	apierrors "github.com/fromforgesoftware/go-kit/errors"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// ChannelSender delivers a notification over a single transport. Concrete
// senders (SMTP today; SES/Mailgun later) share this seam.
type ChannelSender interface {
	Channel() string
	Send(ctx context.Context, n domain.Notification) error
}

// ConfirmingSender is an optional capability a ChannelSender may implement to
// declare what a successful synchronous Send means for the delivery lifecycle.
//
// A transport whose Send only returns nil once the message has been accepted by
// the final delivery server (SMTP, where SendMail returns nil only after the
// receiving MX has accepted the message) confirms delivery and reports
// StatusDelivered. A transport that merely hands the message to an upstream
// gateway that then delivers asynchronously (an HTTP relay returning 2xx for
// SMS/push) only confirms acceptance and reports StatusSent.
//
// Senders that do not implement this default to StatusSent, preserving the
// historical "handed off, not confirmed" semantics.
type ConfirmingSender interface {
	// SuccessStatus is the terminal status to record when Send returns nil.
	SuccessStatus() domain.NotificationStatus
}

// successStatus reports the status to record for a confirmed (error-free) send:
// the sender's declared status when it implements ConfirmingSender, otherwise
// StatusSent.
func successStatus(s ChannelSender) domain.NotificationStatus {
	if cs, ok := s.(ConfirmingSender); ok {
		if st := cs.SuccessStatus(); st.Valid() {
			return st
		}
	}
	return domain.StatusSent
}

// ChannelRegistry resolves a ChannelSender by its channel name.
type ChannelRegistry struct {
	senders map[string]ChannelSender
}

func NewChannelRegistry(senders ...ChannelSender) *ChannelRegistry {
	r := &ChannelRegistry{senders: make(map[string]ChannelSender, len(senders))}
	for _, s := range senders {
		if s != nil {
			r.senders[s.Channel()] = s
		}
	}
	return r
}

func (r *ChannelRegistry) Sender(channel string) (ChannelSender, error) {
	s, ok := r.senders[channel]
	if !ok {
		return nil, apierrors.InvalidArgument("unsupported channel: " + channel)
	}
	return s, nil
}
