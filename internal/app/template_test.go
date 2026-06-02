package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/app/apptest"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

func TestRenderTemplate_InterpolatesData(t *testing.T) {
	subject, body, err := app.RenderTemplate("Hi {{.name}}", "You have {{.count}} messages", map[string]any{
		"name": "alice", "count": 3,
	})
	require.NoError(t, err)
	assert.Equal(t, "Hi alice", subject)
	assert.Equal(t, "You have 3 messages", body)
}

func TestTemplateUsecase_RenderResolvesAndInterpolates(t *testing.T) {
	repo := apptest.NewNotificationTemplateRepository(t)
	uc := app.NewTemplateUsecase(repo)
	repo.EXPECT().GetForRender(mock.Anything, "r1", "welcome", "es").Return(
		domain.NewNotificationTemplate("r1", "welcome",
			domain.WithTemplateLocale("es"),
			domain.WithTemplateSubject("Hola {{.name}}"),
			domain.WithTemplateBody("Bienvenida")), nil)

	subject, body, err := uc.Render(context.Background(), "r1", "welcome", "es", map[string]any{"name": "alice"})
	require.NoError(t, err)
	assert.Equal(t, "Hola alice", subject)
	assert.Equal(t, "Bienvenida", body)
}

func TestSend_RendersTemplateBeforeDispatch(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	sender.EXPECT().Send(mock.Anything, mock.Anything).Return(nil)
	renderer := apptest.NewTemplateRenderer(t)
	renderer.EXPECT().Render(mock.Anything, "r1", "welcome", "es", mock.Anything).Return("Hola", "Bienvenida", nil)

	notifications := apptest.NewNotificationRepository(t)
	attempts := apptest.NewDeliveryAttemptRepository(t)
	uc := app.NewNotificationUsecase(notifications, attempts, app.NewChannelRegistry(sender),
		app.WithSleeper(noSleep), app.WithRenderer(renderer))

	// The persisted notification carries the rendered subject/body.
	notifications.EXPECT().Create(mock.Anything, mock.MatchedBy(func(n domain.Notification) bool {
		return n.Subject() == "Hola" && n.Body() == "Bienvenida"
	})).Return(internaltest.NewNotification(internaltest.WithNID("n-9")), nil)
	attempts.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, nil)
	notifications.EXPECT().UpdateStatus(mock.Anything, "n-9", domain.StatusSent, "").Return(nil)
	notifications.EXPECT().Get(mock.Anything, mock.Anything).
		Return(internaltest.NewNotification(internaltest.WithNID("n-9")), nil)

	n := domain.NewNotification("alice@x.com", domain.ChannelEmail,
		domain.WithNotificationRealmID("r1"),
		domain.WithNotificationTemplate("welcome"),
		domain.WithNotificationLocale("es"),
		domain.WithNotificationData(map[string]any{"name": "alice"}))
	_, err := uc.Send(context.Background(), n)
	require.NoError(t, err)
}
