package app_test

import (
	"context"
	"testing"
	"time"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/app/apptest"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

func usecaseWithPreferences(t *testing.T, sender app.ChannelSender, prefs app.NotificationPreferenceRepository) (app.NotificationUsecase, *apptest.NotificationRepository, *apptest.DeliveryAttemptRepository) {
	t.Helper()
	notifications := apptest.NewNotificationRepository(t)
	attempts := apptest.NewDeliveryAttemptRepository(t)
	uc := app.NewNotificationUsecase(notifications, attempts, app.NewChannelRegistry(sender),
		app.WithPreferences(prefs),
		app.WithSleeper(noSleep),
		app.WithClock(func() time.Time { return time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC) }),
	)
	return uc, notifications, attempts
}

func TestSend_SuppressedChannelPersistsAndSkipsDispatch(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	// sender.Send must NOT be called — mockery fails the test if it is.

	prefs := apptest.NewNotificationPreferenceRepository(t)
	prefs.EXPECT().IsSuppressed(mock.Anything, "", "user@x.com", domain.ChannelEmail).Return(true, nil)

	uc, notifications, _ := usecaseWithPreferences(t, sender, prefs)
	notifications.EXPECT().Create(mock.Anything, mock.MatchedBy(func(n domain.Notification) bool {
		return n.Status() == domain.StatusSuppressed
	})).Return(internaltest.NewNotification(internaltest.WithNID("n-1"), internaltest.WithNStatus(domain.StatusSuppressed)), nil)

	got, err := uc.Send(context.Background(), internaltest.NewNotification(internaltest.WithNRecipient("user@x.com")))
	require.NoError(t, err)
	assert.Equal(t, domain.StatusSuppressed, got.Status())
}

func TestSend_NotSuppressedDispatchesNormally(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	sender.EXPECT().Send(mock.Anything, mock.Anything).Return(nil)

	prefs := apptest.NewNotificationPreferenceRepository(t)
	prefs.EXPECT().IsSuppressed(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

	uc, notifications, attempts := usecaseWithPreferences(t, sender, prefs)
	attempts.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, nil)
	notifications.EXPECT().Create(mock.Anything, mock.Anything).
		Return(internaltest.NewNotification(internaltest.WithNID("n-1")), nil)
	notifications.EXPECT().UpdateStatus(mock.Anything, "n-1", domain.StatusSent, "").Return(nil)
	notifications.EXPECT().Get(mock.Anything, mock.Anything).
		Return(internaltest.NewNotification(internaltest.WithNID("n-1"), internaltest.WithNStatus(domain.StatusSent)), nil)

	got, err := uc.Send(context.Background(), internaltest.NewNotification())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusSent, got.Status())
}

func TestPreferenceUsecase_SetValidates(t *testing.T) {
	prefs := apptest.NewNotificationPreferenceRepository(t)
	uc := app.NewPreferenceUsecase(prefs)

	err := uc.Set(context.Background(), "r", "", domain.ChannelEmail, true)
	assert.True(t, apierrors.Is(err, apierrors.CodeInvalidArgument))
}

func TestPreferenceUsecase_SetUpserts(t *testing.T) {
	prefs := apptest.NewNotificationPreferenceRepository(t)
	prefs.EXPECT().Set(mock.Anything, "r", "user@x.com", domain.ChannelEmail, true).Return(nil)
	uc := app.NewPreferenceUsecase(prefs)

	require.NoError(t, uc.Set(context.Background(), "r", "user@x.com", domain.ChannelEmail, true))
}
