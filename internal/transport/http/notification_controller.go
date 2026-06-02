// Package http holds Gjallarhorn's JSON:API controllers.
package http

import (
	"context"
	"net/http"

	"github.com/fromforgesoftware/go-kit/openapi"
	"github.com/fromforgesoftware/go-kit/search/query"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/gjallarhorn/internal/api"
	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// NotificationController exposes /api/notifications: a send operation plus the
// read surface (list with filters, get-by-id).
type NotificationController struct {
	notifications app.NotificationUsecase
}

func NewNotificationController(notifications app.NotificationUsecase) kitrest.Controller {
	return &NotificationController{notifications: notifications}
}

func (c *NotificationController) Routes(r kitrest.Router) {
	r.Route("/api/notifications", func(r kitrest.Router) {
		r.Post("", kitrest.NewJsonApiCommandHandler(
			c.send, decodeSend, api.NotificationToDTO,
			kitrest.HandlerWithOpenAPI(
				openapi.Summary("Send or schedule a notification"),
				openapi.Description("Dispatches a notification immediately, or schedules it when scheduledAt is set; returns its delivery status."),
				openapi.Tags("notifications"), openapi.Errors(400),
			),
		))
		r.Get("", kitrest.NewJsonApiListHandler(
			c.notifications, api.NotificationToDTO,
			kitrest.HandlerWithOpenAPI(
				openapi.Summary("List notifications"),
				openapi.Description("Filter with filter[status] and filter[recipient]."),
				openapi.Tags("notifications"),
			),
		))
		r.Get("/{id}", kitrest.NewJsonApiGetHandler(
			c.notifications, api.NotificationToDTO, []query.ParseOpt{},
			kitrest.HandlerWithOpenAPI(openapi.Summary("Get a notification"), openapi.Tags("notifications"), openapi.Errors(404)),
		))
	})
}

func (c *NotificationController) send(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	if at := n.ScheduledAt(); at != nil {
		return c.notifications.Schedule(ctx, n, *at)
	}
	return c.notifications.Send(ctx, n)
}

func decodeSend(req *http.Request) (domain.Notification, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.NotificationDTO](req)
	if err != nil {
		return nil, err
	}
	return api.NotificationFromDTO(body), nil
}
