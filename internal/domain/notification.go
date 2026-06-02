// Package domain holds Gjallarhorn's notification-delivery aggregates.
package domain

import (
	"time"

	"github.com/fromforgesoftware/go-kit/resource"
)

const ResourceTypeNotification resource.Type = "notifications"

const (
	ChannelEmail   = "EMAIL"
	ChannelWebhook = "WEBHOOK"
	ChannelSMS     = "SMS"
	ChannelPush    = "PUSH"
)

type NotificationStatus string

// Notification lifecycle states.
//
//   - SCHEDULED  — persisted for a future send; the dispatcher claims it when due.
//   - QUEUED     — accepted, awaiting its first send attempt.
//   - SENT       — handed to an upstream gateway that delivers asynchronously
//     (HTTP relay 2xx for SMS/PUSH/WEBHOOK); final delivery unconfirmed.
//   - DELIVERED  — confirmed accepted by the final delivery server on a
//     synchronous send (SMTP SendMail returns nil only after the receiving MX
//     accepts). Set by the send path when the channel implements
//     app.ConfirmingSender; see internal/app/channel.go.
//   - BOUNCED    — the provider asynchronously reported the message undeliverable
//     after acceptance (hard bounce / rejected delivery). This is inherently
//     provider-dependent: it arrives out-of-band via a provider delivery-status
//     webhook or polling, which this service does not yet ingest. The value is
//     reserved as the documented extension point — when an inbound
//     delivery-status receiver is added it transitions DELIVERED/SENT → BOUNCED
//     via NotificationRepository.UpdateStatus, the same seam the send path uses.
//     It is NOT produced today; see designDecisions.
//   - FAILED     — every synchronous send attempt errored (after retries).
//   - SUPPRESSED — recipient opted out of the channel; never dispatched.
const (
	StatusScheduled  NotificationStatus = "SCHEDULED"
	StatusQueued     NotificationStatus = "QUEUED"
	StatusSent       NotificationStatus = "SENT"
	StatusDelivered  NotificationStatus = "DELIVERED"
	StatusBounced    NotificationStatus = "BOUNCED"
	StatusFailed     NotificationStatus = "FAILED"
	StatusSuppressed NotificationStatus = "SUPPRESSED"
)

func (s NotificationStatus) Valid() bool {
	switch s {
	case StatusScheduled, StatusQueued, StatusSent, StatusDelivered, StatusBounced, StatusFailed, StatusSuppressed:
		return true
	default:
		return false
	}
}

func (s NotificationStatus) String() string { return string(s) }

type Notification interface {
	resource.Resource
	RealmID() string
	Recipient() string
	Channel() string
	Template() string
	Locale() string
	Data() map[string]any
	Subject() string
	Body() string
	Status() NotificationStatus
	LastError() string
	// ScheduledAt is the time a scheduled send becomes due; nil for an
	// immediate send.
	ScheduledAt() *time.Time
}

type notification struct {
	resource.Resource

	realmID     string
	recipient   string
	channel     string
	template    string
	locale      string
	data        map[string]any
	subject     string
	body        string
	status      NotificationStatus
	lastError   string
	scheduledAt *time.Time
}

type NotificationOption func(*notification)

func WithNotificationID(id string) NotificationOption {
	return func(n *notification) { n.Resource = resource.Update(n.Resource, resource.WithID(id)) }
}
func WithNotificationRealmID(id string) NotificationOption {
	return func(n *notification) { n.realmID = id }
}
func WithNotificationTemplate(t string) NotificationOption {
	return func(n *notification) { n.template = t }
}
func WithNotificationLocale(l string) NotificationOption {
	return func(n *notification) { n.locale = l }
}
func WithNotificationData(d map[string]any) NotificationOption {
	return func(n *notification) { n.data = d }
}
func WithNotificationSubject(s string) NotificationOption {
	return func(n *notification) { n.subject = s }
}
func WithNotificationBody(b string) NotificationOption {
	return func(n *notification) { n.body = b }
}
func WithNotificationStatus(s NotificationStatus) NotificationOption {
	return func(n *notification) { n.status = s }
}
func WithNotificationLastError(e string) NotificationOption {
	return func(n *notification) { n.lastError = e }
}
func WithNotificationScheduledAt(t *time.Time) NotificationOption {
	return func(n *notification) { n.scheduledAt = t }
}

func NewNotification(recipient, channel string, opts ...NotificationOption) Notification {
	n := &notification{
		Resource:  resource.New(resource.WithType(ResourceTypeNotification)),
		recipient: recipient,
		channel:   channel,
		status:    StatusQueued,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

func (n *notification) RealmID() string            { return n.realmID }
func (n *notification) Recipient() string          { return n.recipient }
func (n *notification) Channel() string            { return n.channel }
func (n *notification) Template() string           { return n.template }
func (n *notification) Locale() string             { return n.locale }
func (n *notification) Data() map[string]any       { return n.data }
func (n *notification) Subject() string            { return n.subject }
func (n *notification) Body() string               { return n.body }
func (n *notification) Status() NotificationStatus { return n.status }
func (n *notification) LastError() string          { return n.lastError }
func (n *notification) ScheduledAt() *time.Time    { return n.scheduledAt }
