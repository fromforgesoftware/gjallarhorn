// Package api holds Gjallarhorn's JSON:API DTOs and their domain mappers.
package api

import (
	"time"

	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

const ResourceTypeNotification resource.Type = "notifications"

type NotificationDTO struct {
	resource.RestDTO

	RRealmID     string         `jsonapi:"attr,realmId,omitempty"`
	RRecipient   string         `jsonapi:"attr,recipient,omitempty"`
	RChannel     string         `jsonapi:"attr,channel,omitempty"`
	RTemplate    string         `jsonapi:"attr,template,omitempty"`
	RLocale      string         `jsonapi:"attr,locale,omitempty"`
	RData        map[string]any `jsonapi:"attr,data,omitempty"`
	RSubject     string         `jsonapi:"attr,subject,omitempty"`
	RBody        string         `jsonapi:"attr,body,omitempty"`
	RStatus      string         `jsonapi:"attr,status,omitempty"`
	RLastError   string         `jsonapi:"attr,lastError,omitempty"`
	RScheduledAt *time.Time     `jsonapi:"attr,scheduledAt,omitempty"`
	RCreatedAt   time.Time      `jsonapi:"attr,createdAt,omitempty"`
	RUpdatedAt   time.Time      `jsonapi:"attr,updatedAt,omitempty"`
}

func NotificationToDTO(n domain.Notification) *NotificationDTO {
	if n == nil {
		return nil
	}
	dto := &NotificationDTO{
		RestDTO:      resource.ToRestDTO(n),
		RRealmID:     n.RealmID(),
		RRecipient:   n.Recipient(),
		RChannel:     n.Channel(),
		RTemplate:    n.Template(),
		RLocale:      n.Locale(),
		RData:        n.Data(),
		RSubject:     n.Subject(),
		RBody:        n.Body(),
		RStatus:      n.Status().String(),
		RLastError:   n.LastError(),
		RScheduledAt: n.ScheduledAt(),
		RCreatedAt:   n.CreatedAt(),
		RUpdatedAt:   n.UpdatedAt(),
	}
	dto.RType = ResourceTypeNotification
	return dto
}

func NotificationFromDTO(dto *NotificationDTO) domain.Notification {
	if dto == nil {
		return nil
	}
	opts := []domain.NotificationOption{
		domain.WithNotificationRealmID(dto.RRealmID),
		domain.WithNotificationTemplate(dto.RTemplate),
		domain.WithNotificationLocale(dto.RLocale),
		domain.WithNotificationData(dto.RData),
		domain.WithNotificationSubject(dto.RSubject),
		domain.WithNotificationBody(dto.RBody),
		domain.WithNotificationScheduledAt(dto.RScheduledAt),
	}
	return domain.NewNotification(dto.RRecipient, dto.RChannel, opts...)
}
