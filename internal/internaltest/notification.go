// Package internaltest holds test builders + matchers shared by Gjallarhorn's unit
// and integration tests.
package internaltest

import (
	"time"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

type NotificationOption func(*notificationOpts)

type notificationOpts struct {
	id          string
	realmID     string
	recipient   string
	channel     string
	template    string
	data        map[string]any
	subject     string
	body        string
	status      domain.NotificationStatus
	lastError   string
	scheduledAt *time.Time
}

func defaultNotificationOptions() []NotificationOption {
	return []NotificationOption{
		WithNRecipient("user@example.com"),
		WithNChannel(domain.ChannelEmail),
		WithNSubject("hi"),
		WithNBody("hello"),
	}
}

func WithNID(id string) NotificationOption      { return func(o *notificationOpts) { o.id = id } }
func WithNRealmID(id string) NotificationOption { return func(o *notificationOpts) { o.realmID = id } }
func WithNRecipient(r string) NotificationOption {
	return func(o *notificationOpts) { o.recipient = r }
}
func WithNChannel(c string) NotificationOption  { return func(o *notificationOpts) { o.channel = c } }
func WithNTemplate(t string) NotificationOption { return func(o *notificationOpts) { o.template = t } }
func WithNSubject(s string) NotificationOption  { return func(o *notificationOpts) { o.subject = s } }
func WithNBody(b string) NotificationOption     { return func(o *notificationOpts) { o.body = b } }
func WithNLastError(e string) NotificationOption {
	return func(o *notificationOpts) { o.lastError = e }
}
func WithNData(d map[string]any) NotificationOption {
	return func(o *notificationOpts) { o.data = d }
}
func WithNStatus(s domain.NotificationStatus) NotificationOption {
	return func(o *notificationOpts) { o.status = s }
}
func WithNScheduledAt(t time.Time) NotificationOption {
	return func(o *notificationOpts) { o.scheduledAt = &t }
}

func NewNotification(opts ...NotificationOption) domain.Notification {
	o := &notificationOpts{}
	for _, opt := range append(defaultNotificationOptions(), opts...) {
		opt(o)
	}
	domainOpts := []domain.NotificationOption{
		domain.WithNotificationRealmID(o.realmID),
		domain.WithNotificationTemplate(o.template),
		domain.WithNotificationData(o.data),
		domain.WithNotificationSubject(o.subject),
		domain.WithNotificationBody(o.body),
	}
	if o.id != "" {
		domainOpts = append(domainOpts, domain.WithNotificationID(o.id))
	}
	if o.status != "" {
		domainOpts = append(domainOpts, domain.WithNotificationStatus(o.status))
	}
	if o.lastError != "" {
		domainOpts = append(domainOpts, domain.WithNotificationLastError(o.lastError))
	}
	if o.scheduledAt != nil {
		domainOpts = append(domainOpts, domain.WithNotificationScheduledAt(o.scheduledAt))
	}
	return domain.NewNotification(o.recipient, o.channel, domainOpts...)
}

// MatchNotification compares recipient + channel + subject + body, ignoring
// id/status so it works against pre-persist aggregates.
func MatchNotification(want domain.Notification) func(domain.Notification) bool {
	return func(got domain.Notification) bool {
		if want == nil {
			return got == nil
		}
		if got == nil {
			return false
		}
		return want.Recipient() == got.Recipient() &&
			want.Channel() == got.Channel() &&
			want.Subject() == got.Subject() &&
			want.Body() == got.Body()
	}
}
