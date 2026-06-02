package api

import "github.com/fromforgesoftware/go-kit/resource"

const ResourceTypeNotificationPreference resource.Type = "notification-preferences"

// NotificationPreferenceDTO sets/echoes a per-recipient channel opt-out.
type NotificationPreferenceDTO struct {
	resource.RestDTO

	RRealmID    string `jsonapi:"attr,realmId,omitempty"`
	RRecipient  string `jsonapi:"attr,recipient"`
	RChannel    string `jsonapi:"attr,channel"`
	RSuppressed bool   `jsonapi:"attr,suppressed"`
}

func NotificationPreferenceEcho(realmID, recipient, channel string, suppressed bool) *NotificationPreferenceDTO {
	dto := &NotificationPreferenceDTO{
		RRealmID: realmID, RRecipient: recipient, RChannel: channel, RSuppressed: suppressed,
	}
	dto.RType = ResourceTypeNotificationPreference
	return dto
}
