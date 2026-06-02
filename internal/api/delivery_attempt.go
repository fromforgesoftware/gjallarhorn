package api

import (
	"time"

	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

const ResourceTypeDeliveryAttempt resource.Type = "delivery-attempts"

// DeliveryAttemptDTO is the JSON:API representation of one delivery attempt in
// the per-notification attempt log (the console "delivery board").
type DeliveryAttemptDTO struct {
	resource.RestDTO

	RNotificationID string    `jsonapi:"attr,notificationId,omitempty"`
	RAttempt        int       `jsonapi:"attr,attempt,omitempty"`
	RStatus         string    `jsonapi:"attr,status,omitempty"`
	RError          string    `jsonapi:"attr,error,omitempty"`
	RAttemptedAt    time.Time `jsonapi:"attr,attemptedAt,omitempty"`
	RCreatedAt      time.Time `jsonapi:"attr,createdAt,omitempty"`
}

func DeliveryAttemptToDTO(a domain.DeliveryAttempt) *DeliveryAttemptDTO {
	if a == nil {
		return nil
	}
	dto := &DeliveryAttemptDTO{
		RestDTO:         resource.ToRestDTO(a),
		RNotificationID: a.NotificationID(),
		RAttempt:        a.Attempt(),
		RStatus:         a.Status().String(),
		RError:          a.Error(),
		RAttemptedAt:    a.AttemptedAt(),
		RCreatedAt:      a.CreatedAt(),
	}
	dto.RType = ResourceTypeDeliveryAttempt
	return dto
}
