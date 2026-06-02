package http

import (
	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/openapi"
	"github.com/fromforgesoftware/go-kit/search/query"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/gjallarhorn/internal/api"
	"github.com/fromforgesoftware/gjallarhorn/internal/app"
)

// NotificationTemplateController exposes /api/notification-templates CRUD.
type NotificationTemplateController struct {
	templates app.TemplateUsecase
}

func NewNotificationTemplateController(templates app.TemplateUsecase) kitrest.Controller {
	return &NotificationTemplateController{templates: templates}
}

func (c *NotificationTemplateController) Routes(r kitrest.Router) {
	r.Route("/api/notification-templates", func(r kitrest.Router) {
		r.Post("", kitrest.NewJsonApiCreateHandler(
			c.templates, api.NotificationTemplateFromDTO, api.NotificationTemplateToDTO,
			kitrest.HandlerWithOpenAPI(openapi.Summary("Create a notification template"), openapi.Tags("templates"), openapi.Errors(400, 409)),
		))
		r.Get("", kitrest.NewJsonApiListHandler(
			c.templates, api.NotificationTemplateToDTO,
			kitrest.HandlerWithOpenAPI(openapi.Summary("List notification templates"), openapi.Tags("templates")),
		))
		r.Route("/{id}", func(r kitrest.Router) {
			r.Get("", kitrest.NewJsonApiGetHandler(
				c.templates, api.NotificationTemplateToDTO, []query.ParseOpt{},
				kitrest.HandlerWithOpenAPI(openapi.Summary("Get a notification template"), openapi.Tags("templates"), openapi.Errors(404)),
			))
			r.Delete("", kitrest.NewJsonApiDeleteHandler(
				c.templates, repository.DeleteTypeHard,
				kitrest.HandlerWithOpenAPI(openapi.Summary("Delete a notification template"), openapi.Tags("templates"), openapi.Errors(404)),
			))
		})
	})
}
