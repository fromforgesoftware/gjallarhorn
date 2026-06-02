package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

const ResourceTypeNotificationTemplate resource.Type = "notification-templates"

type NotificationTemplateDTO struct {
	resource.RestDTO

	RRealmID string `jsonapi:"attr,realmId,omitempty"`
	RName    string `jsonapi:"attr,name,omitempty"`
	RChannel string `jsonapi:"attr,channel,omitempty"`
	RLocale  string `jsonapi:"attr,locale,omitempty"`
	RSubject string `jsonapi:"attr,subject,omitempty"`
	RBody    string `jsonapi:"attr,body,omitempty"`
}

func NotificationTemplateToDTO(t domain.NotificationTemplate) *NotificationTemplateDTO {
	if t == nil {
		return nil
	}
	dto := &NotificationTemplateDTO{
		RestDTO:  resource.ToRestDTO(t),
		RRealmID: t.RealmID(),
		RName:    t.Name(),
		RChannel: t.Channel(),
		RLocale:  t.Locale(),
		RSubject: t.Subject(),
		RBody:    t.Body(),
	}
	dto.RType = ResourceTypeNotificationTemplate
	return dto
}

func NotificationTemplateFromDTO(dto *NotificationTemplateDTO) domain.NotificationTemplate {
	if dto == nil {
		return nil
	}
	opts := []domain.NotificationTemplateOption{
		domain.WithTemplateLocale(dto.RLocale),
		domain.WithTemplateSubject(dto.RSubject),
		domain.WithTemplateBody(dto.RBody),
	}
	if dto.RChannel != "" {
		opts = append(opts, domain.WithTemplateChannel(dto.RChannel))
	}
	return domain.NewNotificationTemplate(dto.RRealmID, dto.RName, opts...)
}
