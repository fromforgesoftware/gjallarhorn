package domain

import (
	"time"

	"github.com/fromforgesoftware/go-kit/resource"
)

const ResourceTypeDeliveryAttempt resource.Type = "delivery-attempts"

type DeliveryAttempt interface {
	resource.Resource
	NotificationID() string
	Attempt() int
	Status() NotificationStatus
	Error() string
	AttemptedAt() time.Time
}

type deliveryAttempt struct {
	resource.Resource

	notificationID string
	attempt        int
	status         NotificationStatus
	err            string
	attemptedAt    time.Time
}

type DeliveryAttemptOption func(*deliveryAttempt)

func WithDeliveryAttemptID(id string) DeliveryAttemptOption {
	return func(a *deliveryAttempt) { a.Resource = resource.Update(a.Resource, resource.WithID(id)) }
}
func WithDeliveryAttemptError(e string) DeliveryAttemptOption {
	return func(a *deliveryAttempt) { a.err = e }
}
func WithDeliveryAttemptAt(t time.Time) DeliveryAttemptOption {
	return func(a *deliveryAttempt) { a.attemptedAt = t }
}

func NewDeliveryAttempt(notificationID string, attempt int, status NotificationStatus, opts ...DeliveryAttemptOption) DeliveryAttempt {
	a := &deliveryAttempt{
		Resource:       resource.New(resource.WithType(ResourceTypeDeliveryAttempt)),
		notificationID: notificationID,
		attempt:        attempt,
		status:         status,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *deliveryAttempt) NotificationID() string     { return a.notificationID }
func (a *deliveryAttempt) Attempt() int               { return a.attempt }
func (a *deliveryAttempt) Status() NotificationStatus { return a.status }
func (a *deliveryAttempt) Error() string              { return a.err }
func (a *deliveryAttempt) AttemptedAt() time.Time     { return a.attemptedAt }
