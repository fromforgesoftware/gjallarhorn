package app_test

import (
	"context"
	"testing"
	"time"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app/apptest"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

// fixedNow matches the clock newUsecase installs.
var fixedNow = time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)

func TestSchedule_PersistsScheduledWithoutDispatch(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)

	uc, notifications, _ := newUsecase(t, sender)

	// Create is the only repo call — no sender.Send, no UpdateStatus.
	notifications.EXPECT().Create(mock.Anything, mock.MatchedBy(func(n domain.Notification) bool {
		return n.Status() == domain.StatusScheduled && n.ScheduledAt() != nil
	})).Return(internaltest.NewNotification(internaltest.WithNID("n-1"),
		internaltest.WithNStatus(domain.StatusScheduled)), nil)

	got, err := uc.Schedule(context.Background(), internaltest.NewNotification(), fixedNow.Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, domain.StatusScheduled, got.Status())
}

func TestSchedule_RejectsPastTime(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	uc, _, _ := newUsecase(t, sender)

	_, err := uc.Schedule(context.Background(), internaltest.NewNotification(), fixedNow.Add(-time.Hour))
	assert.True(t, apierrors.Is(err, apierrors.CodeInvalidArgument))
}

func TestDispatchDue_ClaimsAndSends(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	sender.EXPECT().Send(mock.Anything, mock.Anything).Return(nil)

	uc, notifications, attempts := newUsecase(t, sender)

	due := internaltest.NewNotification(internaltest.WithNID("n-1"), internaltest.WithNStatus(domain.StatusQueued))
	notifications.EXPECT().ClaimDue(mock.Anything, fixedNow, 50).Return([]domain.Notification{due}, nil)
	attempts.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, nil)
	notifications.EXPECT().UpdateStatus(mock.Anything, "n-1", domain.StatusSent, "").Return(nil)
	notifications.EXPECT().Get(mock.Anything, mock.Anything).
		Return(internaltest.NewNotification(internaltest.WithNID("n-1"), internaltest.WithNStatus(domain.StatusSent)), nil)

	n, err := uc.DispatchDue(context.Background(), 50)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestDispatchDue_NothingDue(t *testing.T) {
	sender := apptest.NewChannelSender(t)
	sender.EXPECT().Channel().Return(domain.ChannelEmail)
	uc, notifications, _ := newUsecase(t, sender)

	notifications.EXPECT().ClaimDue(mock.Anything, fixedNow, 50).Return(nil, nil)

	n, err := uc.DispatchDue(context.Background(), 50)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}
