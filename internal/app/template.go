package app

import (
	"bytes"
	"context"
	"text/template"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	apierrors "github.com/fromforgesoftware/go-kit/errors"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// NotificationTemplateRepository persists templates and resolves them for
// rendering with locale fallback.
type NotificationTemplateRepository interface {
	repository.Creator[domain.NotificationTemplate]
	repository.Getter[domain.NotificationTemplate]
	repository.Lister[domain.NotificationTemplate]
	repository.Deleter
	GetForRender(ctx context.Context, realmID, name, locale string) (domain.NotificationTemplate, error)
}

// TemplateRenderer renders a named template for a realm/locale against data.
// notificationUsecase depends on this narrow port; the usecase below satisfies
// it.
type TemplateRenderer interface {
	Render(ctx context.Context, realmID, name, locale string, data map[string]any) (subject, body string, err error)
}

// TemplateUsecase is the admin surface for templates plus rendering.
type TemplateUsecase interface {
	repository.Getter[domain.NotificationTemplate]
	repository.Lister[domain.NotificationTemplate]
	repository.Deleter
	Create(ctx context.Context, t domain.NotificationTemplate) (domain.NotificationTemplate, error)
	Render(ctx context.Context, realmID, name, locale string, data map[string]any) (subject, body string, err error)
}

type templateUsecase struct {
	usecase.Getter[domain.NotificationTemplate]
	usecase.Lister[domain.NotificationTemplate]
	repository.Deleter

	templates NotificationTemplateRepository
}

func NewTemplateUsecase(templates NotificationTemplateRepository) TemplateUsecase {
	return &templateUsecase{
		Getter:    usecase.NewGetter(templates, domain.ResourceTypeNotificationTemplate),
		Lister:    usecase.NewLister[domain.NotificationTemplate](templates),
		Deleter:   usecase.NewDeleter(templates),
		templates: templates,
	}
}

func (uc *templateUsecase) Create(ctx context.Context, t domain.NotificationTemplate) (domain.NotificationTemplate, error) {
	if t.RealmID() == "" || t.Name() == "" {
		return nil, apierrors.InvalidArgument("realm_id and name are required")
	}
	if _, _, err := RenderTemplate(t.Subject(), t.Body(), map[string]any{}); err != nil {
		return nil, apierrors.InvalidArgument("template does not parse: " + err.Error())
	}
	return uc.templates.Create(ctx, t)
}

func (uc *templateUsecase) Render(ctx context.Context, realmID, name, locale string, data map[string]any) (string, string, error) {
	t, err := uc.templates.GetForRender(ctx, realmID, name, locale)
	if err != nil {
		return "", "", err
	}
	return RenderTemplate(t.Subject(), t.Body(), data)
}

// RenderTemplate executes the subject + body Go templates against data. Pure —
// the rendering contract is unit-tested without a database.
func RenderTemplate(subjectTmpl, bodyTmpl string, data map[string]any) (string, string, error) {
	subject, err := execTemplate("subject", subjectTmpl, data)
	if err != nil {
		return "", "", err
	}
	body, err := execTemplate("body", bodyTmpl, data)
	if err != nil {
		return "", "", err
	}
	return subject, body, nil
}

func execTemplate(name, src string, data map[string]any) (string, error) {
	t, err := template.New(name).Option("missingkey=zero").Parse(src)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
