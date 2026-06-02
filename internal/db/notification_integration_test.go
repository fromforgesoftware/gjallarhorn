//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/db"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

func TestNotificationCreateGetUpdateStatus(t *testing.T) {
	client := internaltest.GetDB(t)
	t.Cleanup(func() { internaltest.TruncateTables(t, client) })

	ctx := context.Background()
	repo, err := db.NewNotificationRepository(client)
	require.NoError(t, err)

	created, err := repo.Create(ctx, internaltest.NewNotification(
		internaltest.WithNRecipient("alice@example.com"),
		internaltest.WithNData(map[string]any{"code": "1234"}),
	))
	require.NoError(t, err)
	require.NotEmpty(t, created.ID())
	assert.Equal(t, domain.StatusQueued, created.Status())
	assert.False(t, created.CreatedAt().IsZero())

	t.Run("get by id", func(t *testing.T) {
		got, err := repo.Get(ctx, internaltest.GetByID(created.ID()))
		require.NoError(t, err)
		assert.Equal(t, "alice@example.com", got.Recipient())
		assert.Equal(t, "1234", got.Data()["code"])
	})

	t.Run("update status", func(t *testing.T) {
		require.NoError(t, repo.UpdateStatus(ctx, created.ID(), domain.StatusFailed, "boom"))
		got, err := repo.Get(ctx, internaltest.GetByID(created.ID()))
		require.NoError(t, err)
		assert.Equal(t, domain.StatusFailed, got.Status())
		assert.Equal(t, "boom", got.LastError())
	})

	t.Run("update missing returns not found", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "00000000-0000-0000-0000-000000000000", domain.StatusSent, "")
		require.Error(t, err)
	})
}

func TestNotificationListFilters(t *testing.T) {
	client := internaltest.GetDB(t)
	t.Cleanup(func() { internaltest.TruncateTables(t, client) })

	ctx := context.Background()
	repo, err := db.NewNotificationRepository(client)
	require.NoError(t, err)

	_, err = repo.Create(ctx, internaltest.NewNotification(
		internaltest.WithNRecipient("bob@example.com"), internaltest.WithNStatus(domain.StatusSent)))
	require.NoError(t, err)
	_, err = repo.Create(ctx, internaltest.NewNotification(
		internaltest.WithNRecipient("carol@example.com"), internaltest.WithNStatus(domain.StatusFailed)))
	require.NoError(t, err)

	t.Run("filter by status", func(t *testing.T) {
		list, err := repo.List(ctx, app.NotificationFilter{Status: string(domain.StatusFailed)}.QueryOptions()...)
		require.NoError(t, err)
		require.Len(t, list.Results(), 1)
		assert.Equal(t, "carol@example.com", list.Results()[0].Recipient())
	})

	t.Run("filter by recipient", func(t *testing.T) {
		list, err := repo.List(ctx, app.NotificationFilter{Recipient: "bob@example.com"}.QueryOptions()...)
		require.NoError(t, err)
		require.Len(t, list.Results(), 1)
		assert.Equal(t, domain.StatusSent, list.Results()[0].Status())
	})
}

func TestDeliveryAttemptCreateAndList(t *testing.T) {
	client := internaltest.GetDB(t)
	t.Cleanup(func() { internaltest.TruncateTables(t, client) })

	ctx := context.Background()
	notifications, err := db.NewNotificationRepository(client)
	require.NoError(t, err)
	attempts, err := db.NewDeliveryAttemptRepository(client)
	require.NoError(t, err)

	n, err := notifications.Create(ctx, internaltest.NewNotification())
	require.NoError(t, err)

	at := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	_, err = attempts.Create(ctx, domain.NewDeliveryAttempt(n.ID(), 1, domain.StatusFailed,
		domain.WithDeliveryAttemptError("timeout"), domain.WithDeliveryAttemptAt(at)))
	require.NoError(t, err)
	_, err = attempts.Create(ctx, domain.NewDeliveryAttempt(n.ID(), 2, domain.StatusSent,
		domain.WithDeliveryAttemptAt(at.Add(time.Second))))
	require.NoError(t, err)

	list, err := attempts.ListByNotification(ctx, n.ID())
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, 1, list[0].Attempt())
	assert.Equal(t, domain.StatusFailed, list[0].Status())
	assert.Equal(t, "timeout", list[0].Error())
	assert.Equal(t, 2, list[1].Attempt())
}
