package http

import (
	"github.com/fromforgesoftware/go-kit/openapi"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/gjallarhorn/internal/api"
	"github.com/fromforgesoftware/gjallarhorn/internal/app"
)

// DeliveryAttemptController exposes /api/delivery-attempts: the per-notification
// attempt log that backs the console "delivery board". The feed is always
// scoped to one notification via filter[notificationId].
type DeliveryAttemptController struct {
	attempts app.DeliveryAttemptUsecase
}

func NewDeliveryAttemptController(attempts app.DeliveryAttemptUsecase) kitrest.Controller {
	return &DeliveryAttemptController{attempts: attempts}
}

func (c *DeliveryAttemptController) Routes(r kitrest.Router) {
	r.Route("/api/delivery-attempts", func(r kitrest.Router) {
		r.Get("", kitrest.NewJsonApiListHandler(
			c.attempts, api.DeliveryAttemptToDTO,
			kitrest.HandlerWithOpenAPI(
				openapi.Summary("List delivery attempts for a notification"),
				openapi.Description("Returns the ordered delivery-attempt log for one notification; requires filter[notificationId][eq]=<id>."),
				openapi.Tags("delivery-attempts"), openapi.Errors(400),
			),
		))
	})
}
