package app

import (
	"context"

	"github.com/fromforgesoftware/go-kit/application/repository"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/fields"
)

// DeliveryAttemptRepository records and reads the per-notification attempt log.
type DeliveryAttemptRepository interface {
	repository.Creator[domain.DeliveryAttempt]
	ListByNotification(ctx context.Context, notificationID string) ([]domain.DeliveryAttempt, error)
}

// DeliveryAttemptUsecase reads the delivery-attempt log feeding the console's
// delivery board. It satisfies usecase.Lister so it plugs into the standard
// JSON:API list handler: the required notificationId comes from the route as a
// filter[notificationId] search option.
type DeliveryAttemptUsecase interface {
	List(ctx context.Context, opts ...search.Option) (resource.ListResponse[domain.DeliveryAttempt], error)
}

type deliveryAttemptUsecase struct {
	attempts DeliveryAttemptRepository
}

func NewDeliveryAttemptUsecase(attempts DeliveryAttemptRepository) DeliveryAttemptUsecase {
	return &deliveryAttemptUsecase{attempts: attempts}
}

// List returns the attempts for the notification named by the
// filter[notificationId] search option, oldest attempt first. The filter is
// required: a delivery-attempt feed is always scoped to one notification, so an
// unscoped list is rejected rather than returning the whole table.
func (uc *deliveryAttemptUsecase) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[domain.DeliveryAttempt], error) {
	notificationID := filterValue(search.New(opts...), fields.NotificationID)
	if notificationID == "" {
		return nil, apierrors.InvalidArgument("filter[" + fields.NotificationID + "] is required")
	}
	found, err := uc.attempts.ListByNotification(ctx, notificationID)
	if err != nil {
		return nil, err
	}
	return resource.NewListResponse(found, len(found)), nil
}

// filterValue extracts the string value of an equality filter on field from a
// search, or "" when absent.
func filterValue(s search.Search, field string) string {
	filters := s.Query().Filters()
	if !filters.Exists(field) {
		return ""
	}
	v, _ := filters.Get(field).Value().(string)
	return v
}
