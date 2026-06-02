package app

import (
	"context"

	"github.com/fromforgesoftware/go-kit/application/repository"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// DeliveryAttemptRepository records and reads the per-notification attempt log.
type DeliveryAttemptRepository interface {
	repository.Creator[domain.DeliveryAttempt]
	ListByNotification(ctx context.Context, notificationID string) ([]domain.DeliveryAttempt, error)
}
