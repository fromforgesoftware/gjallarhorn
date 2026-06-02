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
