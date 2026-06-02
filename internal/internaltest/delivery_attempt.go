package internaltest

import (
	"time"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// NewDeliveryAttempt builds a delivery attempt for tests.
func NewDeliveryAttempt(notificationID string, attempt int, status domain.NotificationStatus, opts ...domain.DeliveryAttemptOption) domain.DeliveryAttempt {
	base := []domain.DeliveryAttemptOption{
		domain.WithDeliveryAttemptAt(time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)),
	}
	return domain.NewDeliveryAttempt(notificationID, attempt, status, append(base, opts...)...)
}
