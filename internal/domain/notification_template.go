package domain

import "github.com/fromforgesoftware/go-kit/resource"

const ResourceTypeNotificationTemplate resource.Type = "notification-templates"

// DefaultLocale is the fallback rendered when a request's locale has no
// matching template.
const DefaultLocale = ""

// NotificationTemplate is a per-realm, localizable message template. Subject
// and Body are Go text/templates rendered against a notification's data.
type NotificationTemplate interface {
	resource.Resource
	RealmID() string
	Name() string
	Channel() string
	Locale() string
	Subject() string
	Body() string
}

type notificationTemplate struct {
	resource.Resource

	realmID string
	name    string
	channel string
	locale  string
	subject string
	body    string
}

type NotificationTemplateOption func(*notificationTemplate)

func WithTemplateID(id string) NotificationTemplateOption {
	return func(t *notificationTemplate) { t.Resource = resource.Update(t.Resource, resource.WithID(id)) }
}
func WithTemplateChannel(c string) NotificationTemplateOption {
	return func(t *notificationTemplate) { t.channel = c }
}
func WithTemplateLocale(l string) NotificationTemplateOption {
	return func(t *notificationTemplate) { t.locale = l }
}
func WithTemplateSubject(s string) NotificationTemplateOption {
	return func(t *notificationTemplate) { t.subject = s }
}
func WithTemplateBody(b string) NotificationTemplateOption {
	return func(t *notificationTemplate) { t.body = b }
}

// NewNotificationTemplate builds a template; realmID + name are mandatory,
// channel defaults to EMAIL.
func NewNotificationTemplate(realmID, name string, opts ...NotificationTemplateOption) NotificationTemplate {
	t := &notificationTemplate{
		Resource: resource.New(resource.WithType(ResourceTypeNotificationTemplate)),
		realmID:  realmID,
		name:     name,
		channel:  ChannelEmail,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *notificationTemplate) RealmID() string { return t.realmID }
func (t *notificationTemplate) Name() string    { return t.name }
func (t *notificationTemplate) Channel() string { return t.channel }
func (t *notificationTemplate) Locale() string  { return t.locale }
func (t *notificationTemplate) Subject() string { return t.subject }
func (t *notificationTemplate) Body() string    { return t.body }
