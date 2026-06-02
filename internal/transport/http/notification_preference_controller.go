package http

import (
	"context"
	"net/http"

	"github.com/fromforgesoftware/go-kit/openapi"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/gjallarhorn/internal/api"
	"github.com/fromforgesoftware/gjallarhorn/internal/app"
)

// NotificationPreferenceController exposes /api/notification-preferences: set a
// per-recipient channel opt-out (idempotent upsert).
type NotificationPreferenceController struct {
	prefs app.PreferenceUsecase
}

func NewNotificationPreferenceController(prefs app.PreferenceUsecase) kitrest.Controller {
	return &NotificationPreferenceController{prefs: prefs}
}

func (c *NotificationPreferenceController) Routes(r kitrest.Router) {
	r.Route("/api/notification-preferences", func(r kitrest.Router) {
		r.Post("", kitrest.NewJsonApiCommandHandler(
			c.set, decodeNotificationPreference, identityNotificationPreferenceDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK),
			kitrest.HandlerWithOpenAPI(
				openapi.Summary("Set a notification preference"),
				openapi.Description("Mutes or unmutes a channel for a recipient (suppressed)."),
				openapi.Tags("preferences"), openapi.Errors(400),
			),
		))
	})
}

func (c *NotificationPreferenceController) set(ctx context.Context, in api.NotificationPreferenceDTO) (*api.NotificationPreferenceDTO, error) {
	if err := c.prefs.Set(ctx, in.RRealmID, in.RRecipient, in.RChannel, in.RSuppressed); err != nil {
		return nil, err
	}
	return api.NotificationPreferenceEcho(in.RRealmID, in.RRecipient, in.RChannel, in.RSuppressed), nil
}

func decodeNotificationPreference(req *http.Request) (api.NotificationPreferenceDTO, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.NotificationPreferenceDTO](req)
	if err != nil {
		return api.NotificationPreferenceDTO{}, err
	}
	return *body, nil
}

func identityNotificationPreferenceDTO(dto *api.NotificationPreferenceDTO) *api.NotificationPreferenceDTO {
	return dto
}
