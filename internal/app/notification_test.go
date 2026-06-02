package app_test

import (
	"context"
	"errors"
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

func noSleep(context.Context, time.Duration) error { return nil }

func newUsecase(t *testing.T, sender app.ChannelSender) (app.NotificationUsecase, *apptest.NotificationRepository, *apptest.DeliveryAttemptRepository) {
	t.Helper()
	notifications := apptest.NewNotificationRepository(t)
	attempts := apptest.NewDeliveryAttemptRepository(t)
	registry := app.NewChannelRegistry(sender)
	uc := app.NewNotificationUsecase(notifications, attempts, registry,
		app.WithSleeper(noSleep),
		app.WithClock(func() time.Time { return time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC) }),
		app.WithBackoffPolicy(app.BackoffPolicy{MaxAttempts: 3, Base: time.Millisecond, Max: time.Millisecond}),
	)
	return uc, notifications, attempts
}

func TestSend_DeliversAndMarksSent(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	sender.EXPECT().Send(mock.Anything, mock.MatchedBy(internaltest.MatchNotification(
		internaltest.NewNotification()))).Return(nil)

	uc, notifications, attempts := newUsecase(t, sender)

	stored := internaltest.NewNotification(internaltest.WithNID("n-1"))
	notifications.EXPECT().Create(mock.Anything, mock.MatchedBy(internaltest.MatchNotification(internaltest.NewNotification()))).
		Return(stored, nil)
	attempts.EXPECT().Create(mock.Anything, mock.MatchedBy(func(a domain.DeliveryAttempt) bool {
		return a.NotificationID() == "n-1" && a.Attempt() == 1 && a.Status() == domain.StatusSent
	})).Return(nil, nil)
	notifications.EXPECT().UpdateStatus(mock.Anything, "n-1", domain.StatusSent, "").Return(nil)
	notifications.EXPECT().Get(mock.Anything, mock.Anything).
		Return(internaltest.NewNotification(internaltest.WithNID("n-1"), internaltest.WithNStatus(domain.StatusSent)), nil)

	got, err := uc.Send(context.Background(), internaltest.NewNotification())
	require.NoError(t, err)
	assert.Equal(t, "n-1", got.ID())
	assert.Equal(t, domain.StatusSent, got.Status())
}

// confirmingSender is a ChannelSender that also implements ConfirmingSender,
// declaring a successful send as a confirmed StatusDelivered (SMTP semantics).
type confirmingSender struct {
	channel string
	err     error
}

func (s confirmingSender) Channel() string                                 { return s.channel }
func (s confirmingSender) Send(context.Context, domain.Notification) error { return s.err }
func (s confirmingSender) SuccessStatus() domain.NotificationStatus        { return domain.StatusDelivered }

func TestSend_ConfirmingSenderMarksDelivered(t *testing.T) {
	uc, notifications, attempts := newUsecase(t, confirmingSender{channel: domain.ChannelEmail})

	stored := internaltest.NewNotification(internaltest.WithNID("n-d"))
	notifications.EXPECT().Create(mock.Anything, mock.Anything).Return(stored, nil)
	attempts.EXPECT().Create(mock.Anything, mock.MatchedBy(func(a domain.DeliveryAttempt) bool {
		return a.NotificationID() == "n-d" && a.Attempt() == 1 && a.Status() == domain.StatusDelivered
	})).Return(nil, nil)
	notifications.EXPECT().UpdateStatus(mock.Anything, "n-d", domain.StatusDelivered, "").Return(nil)
	notifications.EXPECT().Get(mock.Anything, mock.Anything).
		Return(internaltest.NewNotification(internaltest.WithNID("n-d"), internaltest.WithNStatus(domain.StatusDelivered)), nil)

	got, err := uc.Send(context.Background(), internaltest.NewNotification())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDelivered, got.Status())
}

func TestSend_RetriesThenFails(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	sender.EXPECT().Send(mock.Anything, mock.Anything).Return(errors.New("smtp down")).Times(3)

	uc, notifications, attempts := newUsecase(t, sender)

	stored := internaltest.NewNotification(internaltest.WithNID("n-2"))
	notifications.EXPECT().Create(mock.Anything, mock.Anything).Return(stored, nil)
	attempts.EXPECT().Create(mock.Anything, mock.MatchedBy(func(a domain.DeliveryAttempt) bool {
		return a.NotificationID() == "n-2" && a.Status() == domain.StatusFailed
	})).Return(nil, nil).Times(3)
	notifications.EXPECT().UpdateStatus(mock.Anything, "n-2", domain.StatusFailed, "smtp down").Return(nil)
	notifications.EXPECT().Get(mock.Anything, mock.Anything).
		Return(internaltest.NewNotification(internaltest.WithNID("n-2"),
			internaltest.WithNStatus(domain.StatusFailed), internaltest.WithNLastError("smtp down")), nil)

	got, err := uc.Send(context.Background(), internaltest.NewNotification())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusFailed, got.Status())
	assert.Equal(t, "smtp down", got.LastError())
}

func TestSend_RejectsMissingRecipient(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	uc, _, _ := newUsecase(t, sender)

	_, err := uc.Send(context.Background(), domain.NewNotification("", domain.ChannelEmail))
	require.Error(t, err)
	assert.True(t, apierrors.Is(err, apierrors.CodeInvalidArgument))
}

func TestSend_RejectsUnknownChannel(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	uc, _, _ := newUsecase(t, sender)

	_, err := uc.Send(context.Background(), domain.NewNotification("user@example.com", "SMS"))
	require.Error(t, err)
	assert.True(t, apierrors.Is(err, apierrors.CodeInvalidArgument))
}

func TestNotificationFilterQueryOptions_DefaultsLimit(t *testing.T) {
	q := app.NotificationFilter{}.QueryOptions()
	require.Len(t, q, 1)
}
